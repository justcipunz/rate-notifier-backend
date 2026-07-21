package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/justcipunz/rate-notifier-backend/internal/models"
)

type fakeRatesStore struct {
	rates      []models.Rate
	history    map[string][]models.RateHistoryPoint
	historyErr error
}

func (f *fakeRatesStore) CreateUser(ctx context.Context, email, passwordHash string) (models.User, error) {
	panic("unused")
}

func (f *fakeRatesStore) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	panic("unused")
}

func (f *fakeRatesStore) GetUserByID(ctx context.Context, id int64) (models.User, error) {
	panic("unused")
}

func (f *fakeRatesStore) ListRates(ctx context.Context) ([]models.Rate, error) {
	return f.rates, nil
}

func (f *fakeRatesStore) ListRateHistory(ctx context.Context, currency string, from time.Time, to time.Time) ([]models.RateHistoryPoint, error) {
	if f.historyErr != nil {
		return nil, f.historyErr
	}
	return f.history[currency], nil
}

func (f *fakeRatesStore) ListTargetsByUser(ctx context.Context, userID int64) ([]models.Target, error) {
	panic("unused")
}

func (f *fakeRatesStore) CreateTarget(ctx context.Context, target models.Target) (models.Target, error) {
	panic("unused")
}

func (f *fakeRatesStore) GetTargetByID(ctx context.Context, id int64) (models.Target, error) {
	panic("unused")
}

func (f *fakeRatesStore) UpdateTarget(ctx context.Context, target models.Target) (models.Target, error) {
	panic("unused")
}

func (f *fakeRatesStore) DeleteTarget(ctx context.Context, id int64) error {
	panic("unused")
}

func (f *fakeRatesStore) ListNotificationsByUser(ctx context.Context, userID int64) ([]models.Notification, error) {
	panic("unused")
}

func (f *fakeRatesStore) MarkNotificationReadByUser(ctx context.Context, userID, id int64) (models.Notification, error) {
	panic("unused")
}

func (f *fakeRatesStore) GetUserSettings(ctx context.Context, userID int64) (models.UserSettings, error) {
	panic("unused")
}

func (f *fakeRatesStore) UpdateUserSettings(ctx context.Context, userID int64, notificationsEnabled bool) (models.UserSettings, error) {
	panic("unused")
}

func TestCurrencyName(t *testing.T) {
	tests := map[string]string{
		"USD": "Доллар США",
		"EUR": "Евро",
		"CNY": "Китайский юань",
		"GBP": "GBP",
	}

	for code, want := range tests {
		if got := currencyName(code); got != want {
			t.Fatalf("currencyName(%q) = %q, want %q", code, got, want)
		}
	}
}

func TestHandleRatesReturnsNilChangeWhenPreviousMissing(t *testing.T) {
	store := &fakeRatesStore{
		rates: []models.Rate{
			{
				Currency:  "USD",
				Name:      "Доллар США",
				Value:     91.25,
				UpdatedAt: time.Date(2026, 7, 14, 10, 0, 0, 0, time.UTC),
			},
		},
	}
	server := &APIServer{
		store:  store,
		logger: log.New(io.Discard, "", 0),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rates", bytes.NewBuffer(nil))
	rr := httptest.NewRecorder()

	server.handleRates(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}

	var resp ratesResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(resp.Rates) != 1 {
		t.Fatalf("unexpected rates count: %d", len(resp.Rates))
	}
	if resp.Rates[0].PreviousValue != nil {
		t.Fatal("expected previous_value to be nil")
	}
	if resp.Rates[0].Change != nil {
		t.Fatal("expected change to be nil")
	}
}

func TestBuildDailyHistoryFillsMissingDays(t *testing.T) {
	start := time.Date(2026, 7, 17, 0, 0, 0, 0, time.UTC)
	raw := []models.RateHistoryPoint{
		{Value: 90, EffectiveAt: time.Date(2026, 7, 17, 0, 0, 0, 0, time.UTC)},
		{Value: 91, EffectiveAt: time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)},
	}

	points := BuildDailyHistory(raw, start, 4)
	if len(points) != 4 {
		t.Fatalf("expected 4 points, got %d", len(points))
	}

	want := []float64{90, 90, 90, 91}
	for i, value := range want {
		if points[i].Value != value {
			t.Fatalf("point %d = %v, want %v", i, points[i].Value, value)
		}
	}
}

func TestBuildDailyHistoryKeepsOrdinaryWeek(t *testing.T) {
	start := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	raw := []models.RateHistoryPoint{
		{Value: 1, EffectiveAt: start},
		{Value: 2, EffectiveAt: start.AddDate(0, 0, 1)},
		{Value: 3, EffectiveAt: start.AddDate(0, 0, 2)},
		{Value: 4, EffectiveAt: start.AddDate(0, 0, 3)},
		{Value: 5, EffectiveAt: start.AddDate(0, 0, 4)},
		{Value: 6, EffectiveAt: start.AddDate(0, 0, 5)},
		{Value: 7, EffectiveAt: start.AddDate(0, 0, 6)},
	}

	points := BuildDailyHistory(raw, start, 7)
	if len(points) != 7 {
		t.Fatalf("expected 7 points, got %d", len(points))
	}
	for i, point := range points {
		want := float64(i + 1)
		if point.Value != want {
			t.Fatalf("point %d = %v, want %v", i, point.Value, want)
		}
	}
}

func TestBuildDailyHistoryReturnsShortAvailableRange(t *testing.T) {
	start := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	raw := []models.RateHistoryPoint{
		{Value: 90, EffectiveAt: time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC)},
		{Value: 91, EffectiveAt: time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)},
	}

	points := BuildDailyHistory(raw, start, 7)
	if len(points) != 4 {
		t.Fatalf("expected 4 points, got %d", len(points))
	}
	if !points[0].EffectiveAt.Equal(time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected first point date: %v", points[0].EffectiveAt)
	}
	if points[0].Value != 90 || points[3].Value != 91 {
		t.Fatalf("unexpected values: %+v", points)
	}
}

func TestBuildDailyHistoryReturnsFullRangeWithInitialPoint(t *testing.T) {
	start := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	raw := []models.RateHistoryPoint{
		{Value: 89, EffectiveAt: time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)},
		{Value: 91, EffectiveAt: time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)},
	}

	points := BuildDailyHistory(raw, start, 7)
	if len(points) != 7 {
		t.Fatalf("expected 7 points, got %d", len(points))
	}
	if points[0].Value != 89 || points[5].Value != 91 {
		t.Fatalf("unexpected filled values: %+v", points)
	}
}

func TestBuildDailyHistoryReturnsEmptyForNoHistory(t *testing.T) {
	start := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	if points := BuildDailyHistory(nil, start, 7); len(points) != 0 {
		t.Fatalf("expected empty points, got %+v", points)
	}
}

func TestBuildRateHistorySeriesCalculatesTotals(t *testing.T) {
	start := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	raw := []models.RateHistoryPoint{
		{Value: 100, EffectiveAt: start},
		{Value: 110, EffectiveAt: start.AddDate(0, 0, 6)},
	}

	series, ok := BuildRateHistorySeries("USD", raw, start, 7)
	if !ok {
		t.Fatal("expected series")
	}
	if series.StartValue != 100 {
		t.Fatalf("unexpected start value: %v", series.StartValue)
	}
	if series.CurrentValue != 110 {
		t.Fatalf("unexpected current value: %v", series.CurrentValue)
	}
	if series.Change != 10 {
		t.Fatalf("unexpected change: %v", series.Change)
	}
	if series.ChangePercent != 10 {
		t.Fatalf("unexpected change percent: %v", series.ChangePercent)
	}
}

func TestHandleRateHistoryReturnsSeriesInFixedCurrencyOrder(t *testing.T) {
	start := startOfUTCDay(time.Now()).AddDate(0, 0, -(weeklyHistoryDays - 1))
	store := &fakeRatesStore{
		history: map[string][]models.RateHistoryPoint{
			"USD": {{Value: 90, EffectiveAt: start}},
			"EUR": {{Value: 100, EffectiveAt: start}},
			"CNY": {{Value: 12, EffectiveAt: start}},
		},
	}
	server := &APIServer{
		store:  store,
		logger: log.New(io.Discard, "", 0),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rates/history", nil)
	rr := httptest.NewRecorder()

	server.handleRateHistory(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}

	var resp models.RateHistoryResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Period != "7d" {
		t.Fatalf("unexpected period: %s", resp.Period)
	}
	if len(resp.Series) != 3 {
		t.Fatalf("unexpected series count: %d", len(resp.Series))
	}

	want := []string{"USD", "EUR", "CNY"}
	for i, currency := range want {
		if resp.Series[i].Currency != currency {
			t.Fatalf("series[%d] currency = %s, want %s", i, resp.Series[i].Currency, currency)
		}
	}
}

func TestHandleRateHistoryRejectsUnsupportedMethod(t *testing.T) {
	server := &APIServer{
		store:  &fakeRatesStore{},
		logger: log.New(io.Discard, "", 0),
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/rates/history", nil)
	rr := httptest.NewRecorder()

	server.handleRateHistory(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
}

func TestHandleRateHistoryReturnsInternalErrorOnStoreFailure(t *testing.T) {
	server := &APIServer{
		store:  &fakeRatesStore{historyErr: errors.New("db failed")},
		logger: log.New(io.Discard, "", 0),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rates/history", nil)
	rr := httptest.NewRecorder()

	server.handleRateHistory(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
}

func TestHandleRateHistoryReturnsEmptySeriesWhenHistoryMissing(t *testing.T) {
	server := &APIServer{
		store:  &fakeRatesStore{history: map[string][]models.RateHistoryPoint{}},
		logger: log.New(io.Discard, "", 0),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rates/history", nil)
	rr := httptest.NewRecorder()

	server.handleRateHistory(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}

	var resp models.RateHistoryResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(resp.Series) != 0 {
		t.Fatalf("expected empty series, got %+v", resp.Series)
	}
}

func TestServeMuxRoutesRateHistorySeparatelyFromRates(t *testing.T) {
	start := startOfUTCDay(time.Now()).AddDate(0, 0, -(weeklyHistoryDays - 1))
	store := &fakeRatesStore{
		history: map[string][]models.RateHistoryPoint{
			"USD": {{Value: 90, EffectiveAt: start}},
		},
	}
	server := &APIServer{
		store:  store,
		logger: log.New(io.Discard, "", 0),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/rates/history", server.handleRateHistory)
	mux.HandleFunc("/api/v1/rates", server.handleRates)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rates/history", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}

	var resp models.RateHistoryResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Period != "7d" {
		t.Fatalf("expected history response, got period %q", resp.Period)
	}
}

package app

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/justcipunz/rate-notifier-backend/internal/models"
)

type fakeRatesStore struct {
	rates []models.Rate
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

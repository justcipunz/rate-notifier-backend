package app

import (
	"errors"
	"net/http"
	"sort"
	"time"

	"github.com/justcipunz/rate-notifier-backend/internal/httpx"
	"github.com/justcipunz/rate-notifier-backend/internal/models"
	"github.com/justcipunz/rate-notifier-backend/internal/storage"
)

const weeklyHistoryDays = 7

var supportedCurrencies = []string{"USD", "EUR", "CNY"}

type ratesResponse struct {
	Rates []rateDTO `json:"rates"`
}

type rateDTO struct {
	Currency      string   `json:"currency"`
	Name          string   `json:"name"`
	Value         float64  `json:"value"`
	PreviousValue *float64 `json:"previous_value"`
	Change        *float64 `json:"change"`
	UpdatedAt     string   `json:"updated_at"`
}

func (s *APIServer) handleRates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", messageMethodNotAllowed)
		return
	}

	rates, err := s.store.ListRates(r.Context())
	if err != nil {
		s.logInternal("list rates: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", messageInternalError)
		return
	}

	items := make([]rateDTO, 0, len(rates))
	for _, rate := range rates {
		var change *float64
		if rate.PreviousValue != nil {
			value := rate.Value - *rate.PreviousValue
			change = &value
		}

		dto := rateDTO{
			Currency:  rate.Currency,
			Name:      currencyName(rate.Currency),
			Value:     rate.Value,
			Change:    change,
			UpdatedAt: rate.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		}
		dto.PreviousValue = rate.PreviousValue
		items = append(items, dto)
	}

	httpx.WriteJSON(w, http.StatusOK, ratesResponse{Rates: items})
}

func (s *APIServer) handleRateHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", messageMethodNotAllowed)
		return
	}

	series := make([]models.RateHistorySeries, 0, len(supportedCurrencies))
	for _, currency := range supportedCurrencies {
		latest, err := s.store.GetLatestRateHistoryEffectiveAt(r.Context(), currency)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				continue
			}
			s.logInternal("failed to load latest history point: currency=%s error=%v", currency, err)
			httpx.WriteError(w, http.StatusInternalServerError, "internal_error", messageInternalError)
			return
		}

		latestDay := startOfUTCDay(latest)
		start := latestDay.AddDate(0, 0, -(weeklyHistoryDays - 1))
		to := latestDay.AddDate(0, 0, 1).Add(-time.Nanosecond)

		raw, err := s.store.ListRateHistory(r.Context(), currency, start, to)
		if err != nil {
			s.logInternal("failed to load weekly history: currency=%s error=%v", currency, err)
			httpx.WriteError(w, http.StatusInternalServerError, "internal_error", messageInternalError)
			return
		}

		item, ok := BuildRateHistorySeries(currency, raw, start, weeklyHistoryDays)
		if !ok {
			continue
		}
		series = append(series, item)
	}

	httpx.WriteJSON(w, http.StatusOK, models.RateHistoryResponse{
		Period: "7d",
		Series: series,
	})
}

func BuildRateHistorySeries(currency string, raw []models.RateHistoryPoint, start time.Time, days int) (models.RateHistorySeries, bool) {
	points := BuildDailyHistory(raw, start, days)
	if len(points) == 0 {
		return models.RateHistorySeries{}, false
	}

	startValue := points[0].Value
	currentValue := points[len(points)-1].Value
	change := currentValue - startValue

	return models.RateHistorySeries{
		Currency:      currency,
		CurrentValue:  currentValue,
		StartValue:    startValue,
		Change:        change,
		ChangePercent: change / startValue * 100,
		Points:        points,
	}, true
}

func BuildDailyHistory(raw []models.RateHistoryPoint, start time.Time, days int) []models.RateHistoryPoint {
	if len(raw) == 0 || days <= 0 {
		return nil
	}

	start = startOfUTCDay(start)
	end := start.AddDate(0, 0, days)

	items := append([]models.RateHistoryPoint(nil), raw...)
	sort.Slice(items, func(i, j int) bool {
		return items[i].EffectiveAt.Before(items[j].EffectiveAt)
	})

	currentDay := start
	firstDay := startOfUTCDay(items[0].EffectiveAt)
	if firstDay.After(start) {
		currentDay = firstDay
	}
	if !currentDay.Before(end) {
		return nil
	}

	points := make([]models.RateHistoryPoint, 0, days)
	var (
		lastValue float64
		hasValue  bool
		index     int
	)

	for day := currentDay; day.Before(end); day = day.AddDate(0, 0, 1) {
		for index < len(items) {
			itemDay := startOfUTCDay(items[index].EffectiveAt)
			if itemDay.After(day) {
				break
			}
			lastValue = items[index].Value
			hasValue = true
			index++
		}

		if !hasValue {
			continue
		}

		points = append(points, models.RateHistoryPoint{
			Value:       lastValue,
			EffectiveAt: day,
		})
	}

	return points
}

func startOfUTCDay(value time.Time) time.Time {
	utc := value.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}

func currencyName(code string) string {
	switch code {
	case "USD":
		return "Доллар США"
	case "EUR":
		return "Евро"
	case "CNY":
		return "Китайский юань"
	default:
		return code
	}
}

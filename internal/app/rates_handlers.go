package app

import (
	"net/http"

	"github.com/justcipunz/rate-notifier-backend/internal/httpx"
)

type ratesResponse struct {
	Rates []rateDTO `json:"rates"`
}

type rateDTO struct {
	Currency      string   `json:"currency"`
	Name          string   `json:"name"`
	Value         float64  `json:"value"`
	PreviousValue *float64 `json:"previous_value,omitempty"`
	Change        float64  `json:"change"`
	UpdatedAt     string   `json:"updated_at"`
}

func (s *APIServer) handleRates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	rates, err := s.store.ListRates(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "Internal error")
		return
	}

	items := make([]rateDTO, 0, len(rates))
	for _, rate := range rates {
		dto := rateDTO{
			Currency:  rate.Currency,
			Name:      currencyName(rate.Currency),
			Value:     rate.Value,
			Change:    rate.Value - floatOrZero(rate.PreviousValue),
			UpdatedAt: rate.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		}
		if rate.PreviousValue != nil {
			dto.PreviousValue = rate.PreviousValue
		}
		items = append(items, dto)
	}

	httpx.WriteJSON(w, http.StatusOK, ratesResponse{Rates: items})
}

func currencyName(code string) string {
	switch code {
	case "USD":
		return "US Dollar"
	case "EUR":
		return "Euro"
	case "CNY":
		return "Chinese Yuan"
	default:
		return code
	}
}

func floatOrZero(v *float64) float64 {
	if v == nil {
		return 0
	}

	return *v
}

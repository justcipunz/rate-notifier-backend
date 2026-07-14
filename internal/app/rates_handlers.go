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
	PreviousValue *float64 `json:"previous_value"`
	Change        *float64 `json:"change"`
	UpdatedAt     string   `json:"updated_at"`
}

func (s *APIServer) handleRates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	rates, err := s.store.ListRates(r.Context())
	if err != nil {
		s.logInternal("list rates: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "Internal error")
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

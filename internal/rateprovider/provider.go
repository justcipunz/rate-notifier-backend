package rateprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Provider struct {
	url    string
	client *http.Client
}

type Snapshot struct {
	Currency string
	Name     string

	Value         float64
	PreviousValue float64

	EffectiveAt         time.Time
	PreviousEffectiveAt time.Time
}

type cbrResponse struct {
	Date         time.Time              `json:"Date"`
	PreviousDate time.Time              `json:"PreviousDate"`
	Valute       map[string]cbrCurrency `json:"Valute"`
}

type cbrCurrency struct {
	CharCode string  `json:"CharCode"`
	Name     string  `json:"Name"`
	Nominal  float64 `json:"Nominal"`
	Value    float64 `json:"Value"`
	Previous float64 `json:"Previous"`
}

func New(url string, timeout time.Duration) *Provider {
	return &Provider{
		url: url,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (p *Provider) Fetch(ctx context.Context) ([]Snapshot, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch rates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var payload cbrResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode rates response: %w", err)
	}
	if payload.Date.IsZero() {
		return nil, fmt.Errorf("missing rate date")
	}
	if payload.PreviousDate.IsZero() {
		return nil, fmt.Errorf("missing previous rate date")
	}

	rates := make([]Snapshot, 0, 3)
	for _, code := range []string{"USD", "EUR", "CNY"} {
		item, ok := payload.Valute[code]
		if !ok {
			continue
		}

		nominal := item.Nominal
		if nominal <= 0 {
			return nil, fmt.Errorf("invalid nominal for %s", code)
		}

		rates = append(rates, Snapshot{
			Currency:            strings.ToUpper(item.CharCode),
			Name:                item.Name,
			Value:               item.Value / nominal,
			PreviousValue:       item.Previous / nominal,
			EffectiveAt:         payload.Date,
			PreviousEffectiveAt: payload.PreviousDate,
		})
	}

	return rates, nil
}

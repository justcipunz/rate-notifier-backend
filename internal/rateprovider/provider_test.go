package rateprovider

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestProviderFetch_NormalizesValuePreviousAndDates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{
			"Date": "2026-07-21T00:00:00Z",
			"PreviousDate": "2026-07-18T00:00:00Z",
			"Valute": {
				"USD": {"CharCode":"USD","Name":"US Dollar","Nominal":2,"Value":10,"Previous":8},
				"EUR": {"CharCode":"EUR","Name":"Euro","Nominal":1,"Value":11,"Previous":10},
				"CNY": {"CharCode":"CNY","Name":"Chinese Yuan","Nominal":10,"Value":126.7,"Previous":125.1}
			}
		}`)
	}))
	defer server.Close()

	provider := New(server.URL, 0)
	rates, err := provider.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch returned error: %v", err)
	}

	if len(rates) != 3 {
		t.Fatalf("expected 3 rates, got %d", len(rates))
	}

	if rates[0].Currency != "USD" {
		t.Fatalf("unexpected currency: %s", rates[0].Currency)
	}
	if math.Abs(rates[0].Value-5) > 0.0001 {
		t.Fatalf("unexpected USD value: %v", rates[0].Value)
	}
	if math.Abs(rates[0].PreviousValue-4) > 0.0001 {
		t.Fatalf("unexpected USD previous: %v", rates[0].PreviousValue)
	}
	wantDate := time.Date(2026, 7, 21, 0, 0, 0, 0, time.UTC)
	if !rates[0].EffectiveAt.Equal(wantDate) {
		t.Fatalf("unexpected effective date: %v", rates[0].EffectiveAt)
	}
	wantPreviousDate := time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC)
	if !rates[0].PreviousEffectiveAt.Equal(wantPreviousDate) {
		t.Fatalf("unexpected previous effective date: %v", rates[0].PreviousEffectiveAt)
	}

	if rates[2].Currency != "CNY" {
		t.Fatalf("unexpected currency: %s", rates[2].Currency)
	}
	if math.Abs(rates[2].Value-12.67) > 0.0001 {
		t.Fatalf("unexpected CNY value: %v", rates[2].Value)
	}
	if math.Abs(rates[2].PreviousValue-12.51) > 0.0001 {
		t.Fatalf("unexpected CNY previous: %v", rates[2].PreviousValue)
	}
}

func TestProviderFetch_ReturnsErrorOnZeroNominal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{
			"Date": "2026-07-21T00:00:00Z",
			"PreviousDate": "2026-07-18T00:00:00Z",
			"Valute": {
				"USD": {"CharCode":"USD","Name":"US Dollar","Nominal":0,"Value":10,"Previous":9}
			}
		}`)
	}))
	defer server.Close()

	provider := New(server.URL, 0)
	_, err := provider.Fetch(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestProviderFetch_ReturnsErrorOnMissingCurrency(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{
			"Date": "2026-07-21T00:00:00Z",
			"PreviousDate": "2026-07-18T00:00:00Z",
			"Valute": {
				"USD": {"CharCode":"USD","Name":"US Dollar","Nominal":1,"Value":10,"Previous":9}
			}
		}`)
	}))
	defer server.Close()

	provider := New(server.URL, 0)
	_, err := provider.Fetch(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestProviderFetch_ReturnsErrorOnInvalidDate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{
			"Date": "bad-date",
			"PreviousDate": "2026-07-18T00:00:00Z",
			"Valute": {
				"USD": {"CharCode":"USD","Name":"US Dollar","Nominal":1,"Value":10,"Previous":9}
			}
		}`)
	}))
	defer server.Close()

	provider := New(server.URL, 0)
	_, err := provider.Fetch(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestProviderFetch_ReturnsErrorOnMissingDate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{
			"PreviousDate": "2026-07-18T00:00:00Z",
			"Valute": {
				"USD": {"CharCode":"USD","Name":"US Dollar","Nominal":1,"Value":10,"Previous":9}
			}
		}`)
	}))
	defer server.Close()

	provider := New(server.URL, 0)
	_, err := provider.Fetch(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestProviderFetch_ReturnsErrorOnInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{`)
	}))
	defer server.Close()

	provider := New(server.URL, 0)
	_, err := provider.Fetch(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestProviderFetch_ReturnsErrorOnNonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad gateway", http.StatusBadGateway)
	}))
	defer server.Close()

	provider := New(server.URL, 0)
	_, err := provider.Fetch(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

package rateprovider

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProviderFetch_NormalizesValueAndPrevious(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{
			"Valute": {
				"USD": {"CharCode":"USD","Name":"US Dollar","Nominal":2,"Value":10,"Previous":8},
				"CNY": {"CharCode":"CNY","Name":"Chinese Yuan","Nominal":10,"Value":126.7}
			}
		}`)
	}))
	defer server.Close()

	provider := New(server.URL, 0)
	rates, err := provider.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch returned error: %v", err)
	}

	if len(rates) != 2 {
		t.Fatalf("expected 2 rates, got %d", len(rates))
	}

	if rates[0].Currency != "USD" {
		t.Fatalf("unexpected currency: %s", rates[0].Currency)
	}
	if math.Abs(rates[0].Value-5) > 0.0001 {
		t.Fatalf("unexpected USD value: %v", rates[0].Value)
	}
	if rates[0].Previous == nil || math.Abs(*rates[0].Previous-4) > 0.0001 {
		t.Fatalf("unexpected USD previous: %+v", rates[0].Previous)
	}

	if rates[1].Currency != "CNY" {
		t.Fatalf("unexpected currency: %s", rates[1].Currency)
	}
	if math.Abs(rates[1].Value-12.67) > 0.0001 {
		t.Fatalf("unexpected CNY value: %v", rates[1].Value)
	}
	if rates[1].Previous != nil {
		t.Fatalf("expected nil CNY previous, got %+v", rates[1].Previous)
	}
}

func TestProviderFetch_ReturnsErrorOnZeroNominal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{
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

func TestProviderFetch_IgnoresMissingCurrency(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{
			"Valute": {
				"USD": {"CharCode":"USD","Name":"US Dollar","Nominal":1,"Value":10,"Previous":9}
			}
		}`)
	}))
	defer server.Close()

	provider := New(server.URL, 0)
	rates, err := provider.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch returned error: %v", err)
	}

	if len(rates) != 1 {
		t.Fatalf("expected 1 rate, got %d", len(rates))
	}
	if rates[0].Currency != "USD" {
		t.Fatalf("unexpected currency: %s", rates[0].Currency)
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

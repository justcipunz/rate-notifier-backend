package rateprovider

import (
	"context"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProviderFetch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
			"Valute": {
				"USD": {"CharCode":"USD","Name":"Доллар США","Nominal":1,"Value":91.25},
				"EUR": {"CharCode":"EUR","Name":"Евро","Nominal":1,"Value":104.18},
				"CNY": {"CharCode":"CNY","Name":"Китайский юань","Nominal":10,"Value":126.7}
			}
		}`))
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

	if rates[0].Currency != "USD" || math.Abs(rates[0].Value-91.25) > 0.0001 {
		t.Fatalf("unexpected USD rate: %+v", rates[0])
	}

	if rates[2].Currency != "CNY" || math.Abs(rates[2].Value-12.67) > 0.0001 {
		t.Fatalf("unexpected CNY rate: %+v", rates[2])
	}
}

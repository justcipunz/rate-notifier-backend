package app

import "testing"

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

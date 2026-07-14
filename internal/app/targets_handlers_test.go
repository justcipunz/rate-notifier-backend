package app

import (
	"errors"
	"testing"

	"github.com/justcipunz/rate-notifier-backend/internal/models"
)

func TestValidateTargetRequest(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		req := targetRequest{
			Currency:    "usd",
			TargetValue: 95,
			Condition:   "above",
		}

		if err := validateTargetRequest(req); err != nil {
			t.Fatalf("validateTargetRequest returned error: %v", err)
		}
	})

	t.Run("invalid currency", func(t *testing.T) {
		req := targetRequest{
			Currency:    "gbp",
			TargetValue: 95,
			Condition:   "above",
		}

		err := validateTargetRequest(req)
		if !errors.Is(err, errCurrencyNotSupported) {
			t.Fatalf("expected currency error, got: %v", err)
		}
	})

	t.Run("invalid value", func(t *testing.T) {
		req := targetRequest{
			Currency:    "usd",
			TargetValue: 0,
			Condition:   "above",
		}

		if err := validateTargetRequest(req); err == nil {
			t.Fatal("expected validation error for zero value")
		}
	})

	t.Run("invalid condition", func(t *testing.T) {
		req := targetRequest{
			Currency:    "usd",
			TargetValue: 95,
			Condition:   "sideways",
		}

		if err := validateTargetRequest(req); err == nil {
			t.Fatal("expected validation error for unknown condition")
		}
	})
}

func TestTargetMatches(t *testing.T) {
	tests := []struct {
		name   string
		target models.Target
		rate   float64
		want   bool
	}{
		{
			name: "above equal",
			target: models.Target{
				TargetValue: 95,
				Condition:   "above",
			},
			rate: 95,
			want: true,
		},
		{
			name: "above lower",
			target: models.Target{
				TargetValue: 95,
				Condition:   "above",
			},
			rate: 94.9,
			want: false,
		},
		{
			name: "below equal",
			target: models.Target{
				TargetValue: 95,
				Condition:   "below",
			},
			rate: 95,
			want: true,
		},
		{
			name: "below higher",
			target: models.Target{
				TargetValue: 95,
				Condition:   "below",
			},
			rate: 95.1,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := targetMatches(tt.target, tt.rate)
			if got != tt.want {
				t.Fatalf("targetMatches() = %v, want %v", got, tt.want)
			}
		})
	}
}

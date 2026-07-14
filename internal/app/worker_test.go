package app

import (
	"testing"

	"github.com/justcipunz/rate-notifier-backend/internal/models"
)

func TestTargetMatchesAboveAndBelow(t *testing.T) {
	above := models.Target{Condition: "above", TargetValue: 10}
	if !targetMatches(above, 10) {
		t.Fatal("expected above target to match on equality")
	}

	below := models.Target{Condition: "below", TargetValue: 10}
	if !targetMatches(below, 10) {
		t.Fatal("expected below target to match on equality")
	}

	if targetMatches(models.Target{Condition: "unknown", TargetValue: 10}, 10) {
		t.Fatal("expected unknown condition to not match")
	}
}

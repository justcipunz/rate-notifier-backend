package storage

import (
	"strings"
	"testing"
)

func TestListActiveTargetsByCurrencyQueryFiltersDisabledNotifications(t *testing.T) {
	if !strings.Contains(listActiveTargetsByCurrencyQuery, "u.notifications_enabled = TRUE") {
		t.Fatal("expected query to filter by notifications_enabled")
	}
	if !strings.Contains(listActiveTargetsByCurrencyQuery, "JOIN users u ON u.id = t.user_id") {
		t.Fatal("expected query to join users table")
	}
}

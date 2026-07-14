package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/justcipunz/rate-notifier-backend/internal/models"
)

func (s *Store) GetUserSettings(ctx context.Context, userID int64) (models.UserSettings, error) {
	var settings models.UserSettings

	err := s.pool.QueryRow(ctx, `
SELECT notifications_enabled
FROM users
WHERE id = $1`,
		userID,
	).Scan(&settings.NotificationsEnabled)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.UserSettings{}, ErrNotFound
		}
		return models.UserSettings{}, fmt.Errorf("get user settings: %w", err)
	}

	return settings, nil
}

func (s *Store) UpdateUserSettings(ctx context.Context, userID int64, notificationsEnabled bool) (models.UserSettings, error) {
	var settings models.UserSettings

	err := s.pool.QueryRow(ctx, `
UPDATE users
SET notifications_enabled = $2
WHERE id = $1
RETURNING notifications_enabled`,
		userID, notificationsEnabled,
	).Scan(&settings.NotificationsEnabled)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.UserSettings{}, ErrNotFound
		}
		return models.UserSettings{}, fmt.Errorf("update user settings: %w", err)
	}

	return settings, nil
}

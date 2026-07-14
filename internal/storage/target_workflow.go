package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/justcipunz/rate-notifier-backend/internal/models"
)

func (s *Store) TriggerTarget(ctx context.Context, target models.Target, actualValue float64) (models.Notification, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return models.Notification{}, fmt.Errorf("begin target trigger transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var (
		notification models.Notification
		targetID     sql.NullInt64
		createdAt    time.Time
		triggeredAt  sql.NullTime
	)

	err = tx.QueryRow(ctx, `
INSERT INTO notifications (user_id, target_id, currency, target_value, actual_value, condition, is_read, created_at)
VALUES ($1, $2, $3, $4, $5, $6, FALSE, NOW())
RETURNING id, user_id, target_id, currency, target_value, actual_value, condition, is_read, created_at`,
		target.UserID, target.ID, target.Currency, target.TargetValue, actualValue, target.Condition,
	).Scan(&notification.ID, &notification.UserID, &targetID, &notification.Currency, &notification.TargetValue, &notification.ActualValue, &notification.Condition, &notification.IsRead, &createdAt)
	if err != nil {
		return models.Notification{}, fmt.Errorf("create notification in trigger transaction: %w", err)
	}

	notification.TargetID = nullIntPtr(targetID)
	notification.CreatedAt = createdAt

	err = tx.QueryRow(ctx, `
UPDATE targets
SET is_active = FALSE,
    triggered_at = NOW(),
    updated_at = NOW()
WHERE id = $1 AND is_active = TRUE
RETURNING triggered_at`,
		target.ID,
	).Scan(&triggeredAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return models.Notification{}, ErrNotFound
		}
		return models.Notification{}, fmt.Errorf("deactivate triggered target: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return models.Notification{}, fmt.Errorf("commit target trigger transaction: %w", err)
	}

	return notification, nil
}

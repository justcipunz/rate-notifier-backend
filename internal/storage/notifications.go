package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/justcipunz/rate-notifier-backend/internal/models"
)

func (s *Store) CreateNotification(ctx context.Context, notification models.Notification) (models.Notification, error) {
	var (
		created   models.Notification
		targetID  sql.NullInt64
		createdAt time.Time
	)

	err := s.pool.QueryRow(ctx, `
INSERT INTO notifications (user_id, target_id, currency, target_value, actual_value, condition, is_read, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
RETURNING id, user_id, target_id, currency, target_value, actual_value, condition, is_read, created_at`,
		notification.UserID, notification.TargetID, notification.Currency, notification.TargetValue, notification.ActualValue, notification.Condition, notification.IsRead,
	).Scan(&created.ID, &created.UserID, &targetID, &created.Currency, &created.TargetValue, &created.ActualValue, &created.Condition, &created.IsRead, &createdAt)
	if err != nil {
		return models.Notification{}, fmt.Errorf("create notification: %w", err)
	}

	created.TargetID = nullIntPtr(targetID)
	created.CreatedAt = createdAt

	return created, nil
}

func (s *Store) ListNotificationsByUser(ctx context.Context, userID int64) ([]models.Notification, error) {
	rows, err := s.pool.Query(ctx, `
SELECT id, user_id, target_id, currency, target_value, actual_value, condition, is_read, created_at
FROM notifications
WHERE user_id = $1
ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list notifications by user: %w", err)
	}
	defer rows.Close()

	var notifications []models.Notification
	for rows.Next() {
		var (
			n         models.Notification
			targetID  sql.NullInt64
			createdAt time.Time
		)
		if err := rows.Scan(&n.ID, &n.UserID, &targetID, &n.Currency, &n.TargetValue, &n.ActualValue, &n.Condition, &n.IsRead, &createdAt); err != nil {
			return nil, fmt.Errorf("scan notification: %w", err)
		}
		n.TargetID = nullIntPtr(targetID)
		n.CreatedAt = createdAt
		notifications = append(notifications, n)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate notifications: %w", err)
	}

	return notifications, nil
}

func (s *Store) MarkNotificationRead(ctx context.Context, id int64) (models.Notification, error) {
	var (
		n         models.Notification
		targetID  sql.NullInt64
		createdAt time.Time
	)

	err := s.pool.QueryRow(ctx, `
UPDATE notifications
SET is_read = TRUE
WHERE id = $1
RETURNING id, user_id, target_id, currency, target_value, actual_value, condition, is_read, created_at`,
		id,
	).Scan(&n.ID, &n.UserID, &targetID, &n.Currency, &n.TargetValue, &n.ActualValue, &n.Condition, &n.IsRead, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Notification{}, ErrNotFound
		}
		return models.Notification{}, fmt.Errorf("mark notification read: %w", err)
	}

	n.TargetID = nullIntPtr(targetID)
	n.CreatedAt = createdAt

	return n, nil
}

func nullIntPtr(v sql.NullInt64) *int64 {
	if !v.Valid {
		return nil
	}

	value := v.Int64
	return &value
}

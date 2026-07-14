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

func (s *Store) CreateTarget(ctx context.Context, target models.Target) (models.Target, error) {
	var (
		created   models.Target
		triggered sql.NullTime
		createdAt time.Time
		updatedAt time.Time
	)

	err := s.pool.QueryRow(ctx, `
INSERT INTO targets (user_id, currency, target_value, condition, is_active, triggered_at, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
RETURNING id, user_id, currency, target_value, condition, is_active, triggered_at, created_at, updated_at`,
		target.UserID, target.Currency, target.TargetValue, target.Condition, target.IsActive, target.TriggeredAt,
	).Scan(&created.ID, &created.UserID, &created.Currency, &created.TargetValue, &created.Condition, &created.IsActive, &triggered, &createdAt, &updatedAt)
	if err != nil {
		return models.Target{}, fmt.Errorf("create target: %w", err)
	}

	created.TriggeredAt = nullTimePtr(triggered)
	created.CreatedAt = createdAt
	created.UpdatedAt = updatedAt

	return created, nil
}

func (s *Store) ListTargetsByUser(ctx context.Context, userID int64) ([]models.Target, error) {
	rows, err := s.pool.Query(ctx, `
SELECT id, user_id, currency, target_value, condition, is_active, triggered_at, created_at, updated_at
FROM targets
WHERE user_id = $1
ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list targets by user: %w", err)
	}
	defer rows.Close()

	var targets []models.Target
	for rows.Next() {
		var (
			target    models.Target
			triggered sql.NullTime
			createdAt time.Time
			updatedAt time.Time
		)
		if err := rows.Scan(&target.ID, &target.UserID, &target.Currency, &target.TargetValue, &target.Condition, &target.IsActive, &triggered, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan target: %w", err)
		}
		target.TriggeredAt = nullTimePtr(triggered)
		target.CreatedAt = createdAt
		target.UpdatedAt = updatedAt
		targets = append(targets, target)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate targets: %w", err)
	}

	return targets, nil
}

func (s *Store) GetTargetByID(ctx context.Context, id int64) (models.Target, error) {
	var (
		target    models.Target
		triggered sql.NullTime
		createdAt time.Time
		updatedAt time.Time
	)

	err := s.pool.QueryRow(ctx, `
SELECT id, user_id, currency, target_value, condition, is_active, triggered_at, created_at, updated_at
FROM targets
WHERE id = $1`,
		id,
	).Scan(&target.ID, &target.UserID, &target.Currency, &target.TargetValue, &target.Condition, &target.IsActive, &triggered, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Target{}, ErrNotFound
		}
		return models.Target{}, fmt.Errorf("get target by id: %w", err)
	}

	target.TriggeredAt = nullTimePtr(triggered)
	target.CreatedAt = createdAt
	target.UpdatedAt = updatedAt

	return target, nil
}

func (s *Store) UpdateTarget(ctx context.Context, target models.Target) (models.Target, error) {
	var (
		updated   models.Target
		triggered sql.NullTime
		createdAt time.Time
		updatedAt time.Time
	)

	err := s.pool.QueryRow(ctx, `
UPDATE targets
SET currency = $2,
    target_value = $3,
    condition = $4,
    is_active = $5,
    triggered_at = $6,
    updated_at = NOW()
WHERE id = $1
RETURNING id, user_id, currency, target_value, condition, is_active, triggered_at, created_at, updated_at`,
		target.ID, target.Currency, target.TargetValue, target.Condition, target.IsActive, target.TriggeredAt,
	).Scan(&updated.ID, &updated.UserID, &updated.Currency, &updated.TargetValue, &updated.Condition, &updated.IsActive, &triggered, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Target{}, ErrNotFound
		}
		return models.Target{}, fmt.Errorf("update target: %w", err)
	}

	updated.TriggeredAt = nullTimePtr(triggered)
	updated.CreatedAt = createdAt
	updated.UpdatedAt = updatedAt

	return updated, nil
}

func (s *Store) DeleteTarget(ctx context.Context, id int64) error {
	cmd, err := s.pool.Exec(ctx, `DELETE FROM targets WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete target: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) ListActiveTargetsByCurrency(ctx context.Context, currency string) ([]models.Target, error) {
	rows, err := s.pool.Query(ctx, `
SELECT t.id, t.user_id, t.currency, t.target_value, t.condition, t.is_active, t.triggered_at, t.created_at, t.updated_at
FROM targets t
JOIN users u ON u.id = t.user_id
WHERE t.currency = $1
  AND t.is_active = TRUE
  AND u.notifications_enabled = TRUE
ORDER BY t.id`,
		currency,
	)
	if err != nil {
		return nil, fmt.Errorf("list active targets by currency: %w", err)
	}
	defer rows.Close()

	var targets []models.Target
	for rows.Next() {
		var (
			target    models.Target
			triggered sql.NullTime
			createdAt time.Time
			updatedAt time.Time
		)
		if err := rows.Scan(&target.ID, &target.UserID, &target.Currency, &target.TargetValue, &target.Condition, &target.IsActive, &triggered, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan active target: %w", err)
		}
		target.TriggeredAt = nullTimePtr(triggered)
		target.CreatedAt = createdAt
		target.UpdatedAt = updatedAt
		targets = append(targets, target)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate active targets: %w", err)
	}

	return targets, nil
}

func (s *Store) ActivateTarget(ctx context.Context, id int64) (models.Target, error) {
	var (
		target    models.Target
		triggered sql.NullTime
		createdAt time.Time
		updatedAt time.Time
	)

	err := s.pool.QueryRow(ctx, `
UPDATE targets
SET is_active = TRUE,
    triggered_at = NULL,
    updated_at = NOW()
WHERE id = $1
RETURNING id, user_id, currency, target_value, condition, is_active, triggered_at, created_at, updated_at`,
		id,
	).Scan(&target.ID, &target.UserID, &target.Currency, &target.TargetValue, &target.Condition, &target.IsActive, &triggered, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Target{}, ErrNotFound
		}
		return models.Target{}, fmt.Errorf("activate target: %w", err)
	}

	target.TriggeredAt = nullTimePtr(triggered)
	target.CreatedAt = createdAt
	target.UpdatedAt = updatedAt

	return target, nil
}

func (s *Store) DeactivateTarget(ctx context.Context, id int64, triggeredAt *time.Time) (models.Target, error) {
	var (
		target    models.Target
		triggered sql.NullTime
		createdAt time.Time
		updatedAt time.Time
	)

	err := s.pool.QueryRow(ctx, `
UPDATE targets
SET is_active = FALSE,
    triggered_at = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING id, user_id, currency, target_value, condition, is_active, triggered_at, created_at, updated_at`,
		id, triggeredAt,
	).Scan(&target.ID, &target.UserID, &target.Currency, &target.TargetValue, &target.Condition, &target.IsActive, &triggered, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Target{}, ErrNotFound
		}
		return models.Target{}, fmt.Errorf("deactivate target: %w", err)
	}

	target.TriggeredAt = nullTimePtr(triggered)
	target.CreatedAt = createdAt
	target.UpdatedAt = updatedAt

	return target, nil
}

func nullTimePtr(v sql.NullTime) *time.Time {
	if !v.Valid {
		return nil
	}

	value := v.Time
	return &value
}

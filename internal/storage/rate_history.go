package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/justcipunz/rate-notifier-backend/internal/models"
)

func (s *Store) UpsertRateHistory(ctx context.Context, tx pgx.Tx, currency string, value float64, effectiveAt time.Time) error {
	_, err := tx.Exec(ctx, `
INSERT INTO rate_history (
    currency,
    value,
    effective_at,
    fetched_at
)
VALUES ($1, $2, $3, NOW())
ON CONFLICT (currency, effective_at)
DO UPDATE SET
    value = EXCLUDED.value,
    fetched_at = EXCLUDED.fetched_at`,
		currency, value, effectiveAt,
	)
	if err != nil {
		return fmt.Errorf("upsert rate history: %w", err)
	}

	return nil
}

func (s *Store) ListRateHistory(ctx context.Context, currency string, from time.Time, to time.Time) ([]models.RateHistoryPoint, error) {
	rows, err := s.pool.Query(ctx, `
WITH initial_point AS (
    SELECT value, effective_at
    FROM rate_history
    WHERE currency = $1
      AND effective_at <= $2
    ORDER BY effective_at DESC
    LIMIT 1
),
period_points AS (
    SELECT value, effective_at
    FROM rate_history
    WHERE currency = $1
      AND effective_at > $2
      AND effective_at <= $3
)
SELECT value, effective_at
FROM initial_point

UNION ALL

SELECT value, effective_at
FROM period_points

ORDER BY effective_at`,
		currency, from, to,
	)
	if err != nil {
		return nil, fmt.Errorf("list rate history: %w", err)
	}
	defer rows.Close()

	var points []models.RateHistoryPoint
	for rows.Next() {
		var point models.RateHistoryPoint
		if err := rows.Scan(&point.Value, &point.EffectiveAt); err != nil {
			return nil, fmt.Errorf("scan rate history: %w", err)
		}
		points = append(points, point)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rate history: %w", err)
	}

	return points, nil
}

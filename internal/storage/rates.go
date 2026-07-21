package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/justcipunz/rate-notifier-backend/internal/models"
)

func (s *Store) UpsertRate(ctx context.Context, currency string, value float64, previousValue *float64) (models.Rate, error) {
	var (
		rate models.Rate
		prev sql.NullFloat64
	)

	var prevArg any
	if previousValue != nil {
		prevArg = *previousValue
	}

	err := s.pool.QueryRow(ctx, `
INSERT INTO rates (currency, value, previous_value, updated_at)
VALUES ($1, $2, $3, NOW())
ON CONFLICT (currency) DO UPDATE
SET value = EXCLUDED.value,
    previous_value = EXCLUDED.previous_value,
    updated_at = EXCLUDED.updated_at
RETURNING currency, value, previous_value, updated_at`,
		currency, value, prevArg,
	).Scan(&rate.Currency, &rate.Value, &prev, &rate.UpdatedAt)
	if err != nil {
		return models.Rate{}, fmt.Errorf("upsert rate: %w", err)
	}

	rate.PreviousValue = nullFloatPtr(prev)

	return rate, nil
}

func (s *Store) SaveRateSnapshot(
	ctx context.Context,
	currency string,
	value float64,
	previousValue float64,
	effectiveAt time.Time,
	previousEffectiveAt time.Time,
) (models.Rate, int, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return models.Rate{}, 0, fmt.Errorf("begin rate snapshot transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	rate, err := upsertRateTx(ctx, tx, currency, value, previousValue)
	if err != nil {
		return models.Rate{}, 0, err
	}

	if err := s.UpsertRateHistory(ctx, tx, currency, previousValue, previousEffectiveAt); err != nil {
		return models.Rate{}, 0, fmt.Errorf("upsert previous rate history: %w", err)
	}
	if err := s.UpsertRateHistory(ctx, tx, currency, value, effectiveAt); err != nil {
		return models.Rate{}, 0, fmt.Errorf("upsert current rate history: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return models.Rate{}, 0, fmt.Errorf("commit rate snapshot transaction: %w", err)
	}

	return rate, 2, nil
}

func upsertRateTx(ctx context.Context, tx pgx.Tx, currency string, value float64, previousValue float64) (models.Rate, error) {
	var (
		rate models.Rate
		prev sql.NullFloat64
	)

	err := tx.QueryRow(ctx, `
INSERT INTO rates (currency, value, previous_value, updated_at)
VALUES ($1, $2, $3, NOW())
ON CONFLICT (currency) DO UPDATE
SET value = EXCLUDED.value,
    previous_value = EXCLUDED.previous_value,
    updated_at = EXCLUDED.updated_at
RETURNING currency, value, previous_value, updated_at`,
		currency, value, previousValue,
	).Scan(&rate.Currency, &rate.Value, &prev, &rate.UpdatedAt)
	if err != nil {
		return models.Rate{}, fmt.Errorf("upsert rate: %w", err)
	}

	rate.PreviousValue = nullFloatPtr(prev)

	return rate, nil
}

func (s *Store) ListRates(ctx context.Context) ([]models.Rate, error) {
	rows, err := s.pool.Query(ctx, `
SELECT currency, value, previous_value, updated_at
FROM rates
ORDER BY currency`)
	if err != nil {
		return nil, fmt.Errorf("list rates: %w", err)
	}
	defer rows.Close()

	var rates []models.Rate
	for rows.Next() {
		var (
			rate models.Rate
			prev sql.NullFloat64
		)
		if err := rows.Scan(&rate.Currency, &rate.Value, &prev, &rate.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan rate: %w", err)
		}
		rate.PreviousValue = nullFloatPtr(prev)
		rates = append(rates, rate)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rates: %w", err)
	}

	return rates, nil
}

func nullFloatPtr(v sql.NullFloat64) *float64 {
	if !v.Valid {
		return nil
	}

	value := v.Float64
	return &value
}

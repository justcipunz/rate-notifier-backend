package app

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/justcipunz/rate-notifier-backend/internal/config"
	"github.com/justcipunz/rate-notifier-backend/internal/models"
	"github.com/justcipunz/rate-notifier-backend/internal/rateprovider"
	"github.com/justcipunz/rate-notifier-backend/internal/storage"
)

type Worker struct {
	cfg      config.WorkerConfig
	logger   *log.Logger
	db       *pgxpool.Pool
	store    *storage.Store
	provider *rateprovider.Provider
}

func NewWorker(cfg config.WorkerConfig, logger *log.Logger, db *pgxpool.Pool) *Worker {
	return &Worker{
		cfg:      cfg,
		logger:   logger,
		db:       db,
		store:    storage.New(db),
		provider: rateprovider.New(cfg.RateProviderURL, cfg.RateProviderTimeout),
	}
}

func (w *Worker) Run(ctx context.Context) error {
	defer w.db.Close()

	w.logger.Printf("database connected")
	w.logger.Printf("worker started")

	w.syncOnce(ctx)

	ticker := time.NewTicker(w.cfg.RateFetchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Printf("worker stopped")
			return nil
		case <-ticker.C:
			w.syncOnce(ctx)
		}
	}
}

func (w *Worker) syncOnce(ctx context.Context) {
	if err := w.syncRates(ctx); err != nil {
		w.logger.Printf("failed to fetch rates: %v", err)
	}
}

func (w *Worker) syncRates(ctx context.Context) error {
	snapshots, err := w.provider.Fetch(ctx)
	if err != nil {
		return err
	}

	updated := make([]string, 0, len(snapshots))
	for _, snapshot := range snapshots {
		saved, err := w.store.UpsertRate(ctx, snapshot.Currency, snapshot.Value, snapshot.Previous)
		if err != nil {
			return err
		}

		updated = append(updated, saved.Currency)

		if err := w.processTargets(ctx, snapshot.Currency, snapshot.Value); err != nil {
			return err
		}
	}

	if len(updated) > 0 {
		w.logger.Printf("rates updated: %v", updated)
	}

	return nil
}

func (w *Worker) processTargets(ctx context.Context, currency string, currentRate float64) error {
	targets, err := w.store.ListActiveTargetsByCurrency(ctx, currency)
	if err != nil {
		return err
	}

	for _, target := range targets {
		if !targetMatches(target, currentRate) {
			continue
		}

		notification, err := w.store.TriggerTarget(ctx, target, currentRate)
		if err != nil {
			w.logger.Printf("failed to trigger target %d: %v", target.ID, err)
			continue
		}

		w.logger.Printf(
			"target triggered: target_id=%d user_id=%d currency=%s notification_id=%d",
			target.ID,
			target.UserID,
			target.Currency,
			notification.ID,
		)
	}

	return nil
}

func targetMatches(target models.Target, currentRate float64) bool {
	switch target.Condition {
	case "above":
		return currentRate >= target.TargetValue
	case "below":
		return currentRate <= target.TargetValue
	default:
		return false
	}
}

package app

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/justcipunz/rate-notifier-backend/internal/config"
)

type Worker struct {
	cfg    config.Config
	logger *log.Logger
	db     *pgxpool.Pool
}

func NewWorker(cfg config.Config, logger *log.Logger, db *pgxpool.Pool) *Worker {
	return &Worker{
		cfg:    cfg,
		logger: logger,
		db:     db,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	defer w.db.Close()

	w.logger.Printf("database connected")
	w.logger.Printf("worker started")

	<-ctx.Done()

	w.logger.Printf("worker stopped")
	return nil
}

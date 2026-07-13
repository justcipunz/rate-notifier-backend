package app

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/justcipunz/rate-notifier-backend/internal/config"
)

type Worker struct {
	cfg    config.Config
	logger *log.Logger
}

func NewWorker(cfg config.Config, logger *log.Logger) *Worker {
	return &Worker{
		cfg:    cfg,
		logger: logger,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	w.logger.Printf("worker started")

	<-ctx.Done()

	w.logger.Printf("worker stopped")
	return nil
}

func (w *Worker) interval() time.Duration {
	return 5 * time.Minute
}

func (w *Worker) String() string {
	return fmt.Sprintf("worker(port=%s)", w.cfg.AppPort)
}

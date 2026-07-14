package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/justcipunz/rate-notifier-backend/internal/app"
	"github.com/justcipunz/rate-notifier-backend/internal/config"
	"github.com/justcipunz/rate-notifier-backend/internal/db"
	"github.com/justcipunz/rate-notifier-backend/internal/logger"
)

func main() {
	cfg, err := config.LoadWorker()
	if err != nil {
		log.Fatal(err)
	}

	lg := logger.New("worker")

	pool, err := db.Connect(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	worker := app.NewWorker(cfg, lg, pool)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := worker.Run(ctx); err != nil {
		lg.Printf("worker stopped with error: %v", err)
		log.Fatal(err)
	}
}

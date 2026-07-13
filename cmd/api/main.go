package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/justcipunz/rate-notifier-backend/internal/app"
	"github.com/justcipunz/rate-notifier-backend/internal/config"
	"github.com/justcipunz/rate-notifier-backend/internal/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	lg := logger.New("api")

	server := app.NewAPI(cfg, lg)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := server.Run(ctx); err != nil {
		lg.Printf("api stopped with error: %v", err)
		log.Fatal(err)
	}
}

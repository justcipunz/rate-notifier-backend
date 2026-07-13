package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/justcipunz/rate-notifier-backend/internal/config"
)

type APIServer struct {
	cfg    config.Config
	logger *log.Logger
}

func NewAPI(cfg config.Config, logger *log.Logger) *APIServer {
	return &APIServer{
		cfg:    cfg,
		logger: logger,
	}
}

func (s *APIServer) Run(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)

	server := &http.Server{
		Addr:              ":" + s.cfg.AppPort,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		s.logger.Printf("http server started on :%s", s.cfg.AppPort)
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown api: %w", err)
		}
		return nil
	case err := <-errCh:
		if err == nil || errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

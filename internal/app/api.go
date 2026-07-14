package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/justcipunz/rate-notifier-backend/internal/auth"
	"github.com/justcipunz/rate-notifier-backend/internal/config"
	"github.com/justcipunz/rate-notifier-backend/internal/httpx"
	"github.com/justcipunz/rate-notifier-backend/internal/middleware"
	"github.com/justcipunz/rate-notifier-backend/internal/migrations"
	"github.com/justcipunz/rate-notifier-backend/internal/storage"
)

type APIServer struct {
	cfg    config.Config
	logger *log.Logger
	db     *pgxpool.Pool
	store  *storage.Store
	tokens *auth.TokenManager
}

func NewAPI(cfg config.Config, logger *log.Logger, db *pgxpool.Pool) *APIServer {
	return &APIServer{
		cfg:    cfg,
		logger: logger,
		db:     db,
		store:  storage.New(db),
		tokens: auth.NewTokenManager(cfg.JWTSecret, cfg.JWTTTL),
	}
}

func (s *APIServer) Run(ctx context.Context) error {
	defer s.db.Close()

	if err := migrations.Run(ctx, s.db); err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/auth/register", s.handleRegister)
	mux.HandleFunc("/api/v1/auth/login", s.handleLogin)
	mux.HandleFunc("/api/v1/rates", s.handleRates)
	mux.Handle("/api/v1/users/me", middleware.RequireAuth(s.tokens, http.HandlerFunc(s.handleMe)))
	mux.Handle("/api/v1/targets", middleware.RequireAuth(s.tokens, http.HandlerFunc(s.handleTargets)))
	mux.Handle("/api/v1/targets/{id}", middleware.RequireAuth(s.tokens, http.HandlerFunc(s.handleTargetByID)))
	mux.Handle("/api/v1/notifications", middleware.RequireAuth(s.tokens, http.HandlerFunc(s.handleNotifications)))
	mux.Handle("/api/v1/notifications/{id}/read", middleware.RequireAuth(s.tokens, http.HandlerFunc(s.handleNotificationRead)))
	mux.HandleFunc("/health", healthHandler)

	server := &http.Server{
		Addr:              ":" + s.cfg.AppPort,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		s.logger.Printf("database connected")
		s.logger.Printf("migrations completed")
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
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

package middleware

import (
	"log"
	"net/http"

	"github.com/justcipunz/rate-notifier-backend/internal/httpx"
)

func Recovery(logger *log.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Printf("panic recovered: %v", rec)
					httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "Внутренняя ошибка")
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

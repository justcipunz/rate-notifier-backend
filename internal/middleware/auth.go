package middleware

import (
	"net/http"
	"strings"

	"github.com/justcipunz/rate-notifier-backend/internal/auth"
	"github.com/justcipunz/rate-notifier-backend/internal/httpx"
)

func RequireAuth(tm *auth.TokenManager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "Требуется авторизация")
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		claims, err := tm.Parse(token)
		if err != nil {
			httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "Требуется авторизация")
			return
		}

		ctx := auth.WithPrincipal(r.Context(), auth.Principal{
			ID:    claims.UserID,
			Email: claims.Email,
		})

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

package app

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/justcipunz/rate-notifier-backend/internal/auth"
	"github.com/justcipunz/rate-notifier-backend/internal/httpx"
	"github.com/justcipunz/rate-notifier-backend/internal/storage"
)

type settingsRequest struct {
	NotificationsEnabled *bool `json:"notifications_enabled"`
}

type settingsResponse struct {
	NotificationsEnabled bool `json:"notifications_enabled"`
}

func (s *APIServer) handleSettings(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "Требуется авторизация")
		return
	}

	switch r.Method {
	case http.MethodGet:
		settings, err := s.store.GetUserSettings(r.Context(), principal.ID)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				httpx.WriteError(w, http.StatusNotFound, "settings_not_found", "Settings not found")
				return
			}
			s.logger.Printf("get user settings: %v", err)
			httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "Внутренняя ошибка")
			return
		}

		httpx.WriteJSON(w, http.StatusOK, settingsResponse{
			NotificationsEnabled: settings.NotificationsEnabled,
		})
	case http.MethodPut:
		var req settingsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "validation_error", "Некорректные данные запроса")
			return
		}

		if req.NotificationsEnabled == nil {
			httpx.WriteError(w, http.StatusBadRequest, "validation_error", "notifications_enabled is required")
			return
		}

		settings, err := s.store.UpdateUserSettings(r.Context(), principal.ID, *req.NotificationsEnabled)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				httpx.WriteError(w, http.StatusNotFound, "settings_not_found", "Settings not found")
				return
			}
			s.logger.Printf("update user settings: %v", err)
			httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "Внутренняя ошибка")
			return
		}

		httpx.WriteJSON(w, http.StatusOK, settingsResponse{
			NotificationsEnabled: settings.NotificationsEnabled,
		})
	default:
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
	}
}

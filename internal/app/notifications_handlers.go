package app

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/justcipunz/rate-notifier-backend/internal/auth"
	"github.com/justcipunz/rate-notifier-backend/internal/httpx"
	"github.com/justcipunz/rate-notifier-backend/internal/models"
	"github.com/justcipunz/rate-notifier-backend/internal/storage"
)

type notificationDTO struct {
	ID          int64     `json:"id"`
	TargetID    *int64    `json:"target_id,omitempty"`
	Currency    string    `json:"currency"`
	TargetValue float64   `json:"target_value"`
	ActualValue float64   `json:"actual_value"`
	Condition   string    `json:"condition"`
	IsRead      bool      `json:"is_read"`
	CreatedAt   time.Time `json:"created_at"`
}

type notificationsResponse struct {
	Notifications []notificationDTO `json:"notifications"`
}

type notificationReadResponse struct {
	ID     int64 `json:"id"`
	IsRead bool  `json:"is_read"`
}

func (s *APIServer) handleNotifications(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "Authorization required")
		return
	}

	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	notifications, err := s.store.ListNotificationsByUser(r.Context(), principal.ID)
	if err != nil {
		s.logInternal("list notifications: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "Internal error")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, notificationsResponse{Notifications: mapNotifications(notifications)})
}

func (s *APIServer) handleNotificationRead(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "Authorization required")
		return
	}

	if r.Method != http.MethodPut {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "validation_error", "Invalid notification ID")
		return
	}

	notification, err := s.store.MarkNotificationReadByUser(r.Context(), principal.ID, id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			httpx.WriteError(w, http.StatusNotFound, "notification_not_found", "Notification not found")
			return
		}
		s.logInternal("mark notification read: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "Internal error")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, notificationReadResponse{
		ID:     notification.ID,
		IsRead: notification.IsRead,
	})
}

func mapNotifications(items []models.Notification) []notificationDTO {
	result := make([]notificationDTO, 0, len(items))
	for _, item := range items {
		result = append(result, notificationDTO{
			ID:          item.ID,
			TargetID:    item.TargetID,
			Currency:    item.Currency,
			TargetValue: item.TargetValue,
			ActualValue: item.ActualValue,
			Condition:   item.Condition,
			IsRead:      item.IsRead,
			CreatedAt:   item.CreatedAt,
		})
	}
	return result
}

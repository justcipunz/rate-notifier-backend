package app

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/justcipunz/rate-notifier-backend/internal/auth"
	"github.com/justcipunz/rate-notifier-backend/internal/httpx"
	"github.com/justcipunz/rate-notifier-backend/internal/models"
	"github.com/justcipunz/rate-notifier-backend/internal/storage"
)

type targetRequest struct {
	Currency    string  `json:"currency"`
	TargetValue float64 `json:"target_value"`
	Condition   string  `json:"condition"`
}

type targetUpdateRequest struct {
	Currency    string  `json:"currency"`
	TargetValue float64 `json:"target_value"`
	Condition   string  `json:"condition"`
	IsActive    *bool   `json:"is_active"`
}

type targetDTO struct {
	ID          int64      `json:"id"`
	Currency    string     `json:"currency"`
	TargetValue float64    `json:"target_value"`
	Condition   string     `json:"condition"`
	IsActive    bool       `json:"is_active"`
	TriggeredAt *time.Time `json:"triggered_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type targetsResponse struct {
	Targets []targetDTO `json:"targets"`
}

func (s *APIServer) handleTargets(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "Authorization required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		targets, err := s.store.ListTargetsByUser(r.Context(), principal.ID)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "Internal error")
			return
		}

		httpx.WriteJSON(w, http.StatusOK, targetsResponse{Targets: mapTargets(targets)})
	case http.MethodPost:
		s.createTarget(w, r, principal.ID)
	default:
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
	}
}

func (s *APIServer) handleTargetByID(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "Authorization required")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "validation_error", "Invalid target ID")
		return
	}

	switch r.Method {
	case http.MethodPut:
		s.updateTarget(w, r, principal.ID, id)
	case http.MethodDelete:
		s.deleteTarget(w, r, principal.ID, id)
	default:
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
	}
}

func (s *APIServer) createTarget(w http.ResponseWriter, r *http.Request, userID int64) {
	var req targetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "validation_error", "Invalid request data")
		return
	}

	if err := validateTargetFields(req.Currency, req.TargetValue, req.Condition); err != nil {
		code := "validation_error"
		if errors.Is(err, errCurrencyNotSupported) {
			code = "currency_not_supported"
		}
		httpx.WriteError(w, http.StatusBadRequest, code, err.Error())
		return
	}

	created, err := s.store.CreateTarget(r.Context(), models.Target{
		UserID:      userID,
		Currency:    strings.ToUpper(strings.TrimSpace(req.Currency)),
		TargetValue: req.TargetValue,
		Condition:   strings.ToLower(strings.TrimSpace(req.Condition)),
		IsActive:    true,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "Internal error")
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, map[string]any{
		"target": toTargetDTO(created),
	})
}

func (s *APIServer) updateTarget(w http.ResponseWriter, r *http.Request, userID, targetID int64) {
	var req targetUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "validation_error", "Invalid request data")
		return
	}

	if req.IsActive == nil {
		httpx.WriteError(w, http.StatusBadRequest, "validation_error", "is_active is required")
		return
	}

	if err := validateTargetFields(req.Currency, req.TargetValue, req.Condition); err != nil {
		code := "validation_error"
		if errors.Is(err, errCurrencyNotSupported) {
			code = "currency_not_supported"
		}
		httpx.WriteError(w, http.StatusBadRequest, code, err.Error())
		return
	}

	current, err := s.store.GetTargetByID(r.Context(), targetID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			httpx.WriteError(w, http.StatusNotFound, "target_not_found", "Target not found")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "Internal error")
		return
	}

	if current.UserID != userID {
		httpx.WriteError(w, http.StatusNotFound, "target_not_found", "Target not found")
		return
	}

	updated := models.Target{
		ID:          current.ID,
		UserID:      current.UserID,
		Currency:    strings.ToUpper(strings.TrimSpace(req.Currency)),
		TargetValue: req.TargetValue,
		Condition:   strings.ToLower(strings.TrimSpace(req.Condition)),
		IsActive:    *req.IsActive,
	}

	if *req.IsActive {
		updated.TriggeredAt = nil
	} else {
		updated.TriggeredAt = current.TriggeredAt
	}

	saved, err := s.store.UpdateTarget(r.Context(), updated)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			httpx.WriteError(w, http.StatusNotFound, "target_not_found", "Target not found")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "Internal error")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"target": toTargetDTO(saved),
	})
}

func (s *APIServer) deleteTarget(w http.ResponseWriter, r *http.Request, userID, targetID int64) {
	current, err := s.store.GetTargetByID(r.Context(), targetID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			httpx.WriteError(w, http.StatusNotFound, "target_not_found", "Target not found")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "Internal error")
		return
	}

	if current.UserID != userID {
		httpx.WriteError(w, http.StatusNotFound, "target_not_found", "Target not found")
		return
	}

	if err := s.store.DeleteTarget(r.Context(), targetID); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			httpx.WriteError(w, http.StatusNotFound, "target_not_found", "Target not found")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "Internal error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

var errCurrencyNotSupported = errors.New("currency not supported")

func validateTargetFields(currency string, targetValue float64, condition string) error {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	switch currency {
	case "USD", "EUR", "CNY":
	default:
		return errCurrencyNotSupported
	}

	if targetValue <= 0 {
		return errors.New("target value must be greater than zero")
	}

	condition = strings.ToLower(strings.TrimSpace(condition))
	switch condition {
	case "above", "below":
	default:
		return errors.New("unknown condition")
	}

	return nil
}

func mapTargets(targets []models.Target) []targetDTO {
	items := make([]targetDTO, 0, len(targets))
	for _, target := range targets {
		items = append(items, toTargetDTO(target))
	}
	return items
}

func toTargetDTO(target models.Target) targetDTO {
	return targetDTO{
		ID:          target.ID,
		Currency:    target.Currency,
		TargetValue: target.TargetValue,
		Condition:   target.Condition,
		IsActive:    target.IsActive,
		TriggeredAt: target.TriggeredAt,
		CreatedAt:   target.CreatedAt,
		UpdatedAt:   target.UpdatedAt,
	}
}

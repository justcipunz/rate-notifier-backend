package app

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/justcipunz/rate-notifier-backend/internal/auth"
	"github.com/justcipunz/rate-notifier-backend/internal/httpx"
	"github.com/justcipunz/rate-notifier-backend/internal/models"
	"github.com/justcipunz/rate-notifier-backend/internal/storage"
)

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string  `json:"token"`
	User  userDTO `json:"user"`
}

type userDTO struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

func (s *APIServer) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", messageMethodNotAllowed)
		return
	}

	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "validation_error", messageInvalidRequestData)
		return
	}

	email, err := auth.NormalizeEmail(req.Email)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "validation_error", messageInvalidEmail)
		return
	}

	if err := auth.ValidatePassword(req.Password); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		s.logInternal("hash password: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", messageInternalError)
		return
	}

	user, err := s.store.CreateUser(r.Context(), email, passwordHash)
	if err != nil {
		if errors.Is(err, storage.ErrEmailExists) {
			httpx.WriteError(w, http.StatusConflict, "email_already_exists", messageEmailAlreadyExists)
			return
		}
		s.logInternal("create user: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", messageInternalError)
		return
	}

	token, err := s.tokens.Generate(user.ID, user.Email)
	if err != nil {
		s.logInternal("generate token: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", messageInternalError)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, authResponse{
		Token: token,
		User:  userToDTO(user),
	})
}

func (s *APIServer) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", messageMethodNotAllowed)
		return
	}

	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "validation_error", messageInvalidRequestData)
		return
	}

	email, err := auth.NormalizeEmail(req.Email)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "validation_error", messageInvalidEmail)
		return
	}

	user, err := s.store.GetUserByEmail(r.Context(), email)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			httpx.WriteError(w, http.StatusUnauthorized, "invalid_credentials", messageInvalidCredentials)
			return
		}
		s.logInternal("get user by email: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", messageInternalError)
		return
	}

	if err := auth.CheckPassword(user.PasswordHash, req.Password); err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, "invalid_credentials", messageInvalidCredentials)
		return
	}

	token, err := s.tokens.Generate(user.ID, user.Email)
	if err != nil {
		s.logInternal("generate token: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", messageInternalError)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, authResponse{
		Token: token,
		User:  userToDTO(user),
	})
}

func (s *APIServer) handleMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", messageMethodNotAllowed)
		return
	}

	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", messageAuthRequired)
		return
	}

	user, err := s.store.GetUserByID(r.Context(), principal.ID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", messageAuthRequired)
			return
		}
		s.logInternal("get current user: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", messageInternalError)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, userToDTO(user))
}

func userToDTO(user models.User) userDTO {
	return userDTO{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}

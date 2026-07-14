package app

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

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
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "validation_error", "Некорректные данные запроса")
		return
	}

	email, err := auth.NormalizeEmail(req.Email)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "validation_error", "Некорректный email")
		return
	}

	if err := auth.ValidatePassword(req.Password); err != nil {
		message := "Пароль должен содержать не менее 8 байт"
		if strings.Contains(err.Error(), "at most 72 bytes") {
			message = "Пароль должен содержать не более 72 байт"
		}
		httpx.WriteError(w, http.StatusBadRequest, "validation_error", message)
		return
	}

	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		s.logger.Printf("hash password: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "Внутренняя ошибка")
		return
	}

	user, err := s.store.CreateUser(r.Context(), email, passwordHash)
	if err != nil {
		if errors.Is(err, storage.ErrEmailExists) {
			httpx.WriteError(w, http.StatusConflict, "email_already_exists", "Email уже зарегистрирован")
			return
		}
		s.logger.Printf("create user: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "Внутренняя ошибка")
		return
	}

	token, err := s.tokens.Generate(user.ID, user.Email)
	if err != nil {
		s.logger.Printf("generate token: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "Внутренняя ошибка")
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, authResponse{
		Token: token,
		User:  userToDTO(user),
	})
}

func (s *APIServer) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "validation_error", "Некорректные данные запроса")
		return
	}

	email, err := auth.NormalizeEmail(req.Email)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "validation_error", "Некорректный email")
		return
	}

	user, err := s.store.GetUserByEmail(r.Context(), email)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			httpx.WriteError(w, http.StatusUnauthorized, "invalid_credentials", "Неверный email или пароль")
			return
		}
		s.logger.Printf("get user by email: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "Внутренняя ошибка")
		return
	}

	if err := auth.CheckPassword(user.PasswordHash, req.Password); err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, "invalid_credentials", "Неверный email или пароль")
		return
	}

	token, err := s.tokens.Generate(user.ID, user.Email)
	if err != nil {
		s.logger.Printf("generate token: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "Внутренняя ошибка")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, authResponse{
		Token: token,
		User:  userToDTO(user),
	})
}

func (s *APIServer) handleMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "Требуется авторизация")
		return
	}

	user, err := s.store.GetUserByID(r.Context(), principal.ID)
	if err != nil {
		s.logger.Printf("get user by id: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "Внутренняя ошибка")
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

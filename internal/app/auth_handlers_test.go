package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/justcipunz/rate-notifier-backend/internal/auth"
	"github.com/justcipunz/rate-notifier-backend/internal/httpx"
	"github.com/justcipunz/rate-notifier-backend/internal/storage"
)

func TestHandleMeReturnsUnauthorizedWhenUserMissing(t *testing.T) {
	store := &fakeSettingsStore{
		getUserErr: storage.ErrNotFound,
	}
	server := &APIServer{
		store:  store,
		logger: log.New(io.Discard, "", 0),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.Principal{ID: 42, Email: "user@example.com"}))
	rr := httptest.NewRecorder()

	server.handleMe(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if !store.getUserCalled {
		t.Fatal("expected GetUserByID to be called")
	}

	var resp httpx.ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Error.Code != "unauthorized" {
		t.Fatalf("unexpected error code: %s", resp.Error.Code)
	}
}

func TestHandleMeReturnsInternalErrorOnStoreFailure(t *testing.T) {
	store := &fakeSettingsStore{
		getUserErr: errors.New("db failed"),
	}
	server := &APIServer{
		store:  store,
		logger: log.New(io.Discard, "", 0),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.Principal{ID: 42, Email: "user@example.com"}))
	rr := httptest.NewRecorder()

	server.handleMe(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
}

func TestHandleLoginReturnsUnauthorizedWhenUserMissing(t *testing.T) {
	store := &fakeSettingsStore{
		getByEmailErr: storage.ErrNotFound,
	}
	server := &APIServer{
		store:  store,
		logger: log.New(io.Discard, "", 0),
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"email":"user@example.com","password":"password123"}`))
	rr := httptest.NewRecorder()

	server.handleLogin(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if !store.getByEmailCalled {
		t.Fatal("expected GetUserByEmail to be called")
	}

	var resp httpx.ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Error.Code != "invalid_credentials" {
		t.Fatalf("unexpected error code: %s", resp.Error.Code)
	}
}

func TestHandleLoginReturnsInternalErrorOnStoreFailure(t *testing.T) {
	store := &fakeSettingsStore{
		getByEmailErr: errors.New("db failed"),
	}
	server := &APIServer{
		store:  store,
		logger: log.New(io.Discard, "", 0),
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"email":"user@example.com","password":"password123"}`))
	rr := httptest.NewRecorder()

	server.handleLogin(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
}

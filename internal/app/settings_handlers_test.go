package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/justcipunz/rate-notifier-backend/internal/auth"
	"github.com/justcipunz/rate-notifier-backend/internal/models"
)

type fakeSettingsStore struct {
	getSettings  models.UserSettings
	getErr       error
	updateErr    error
	updateCalled bool
	updateValue  bool
	updateUserID int64
	getCalled    bool
	getUserID    int64
}

func (f *fakeSettingsStore) CreateUser(ctx context.Context, email, passwordHash string) (models.User, error) {
	panic("unused")
}

func (f *fakeSettingsStore) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	panic("unused")
}

func (f *fakeSettingsStore) GetUserByID(ctx context.Context, id int64) (models.User, error) {
	panic("unused")
}

func (f *fakeSettingsStore) ListRates(ctx context.Context) ([]models.Rate, error) {
	panic("unused")
}

func (f *fakeSettingsStore) ListTargetsByUser(ctx context.Context, userID int64) ([]models.Target, error) {
	panic("unused")
}

func (f *fakeSettingsStore) CreateTarget(ctx context.Context, target models.Target) (models.Target, error) {
	panic("unused")
}

func (f *fakeSettingsStore) GetTargetByID(ctx context.Context, id int64) (models.Target, error) {
	panic("unused")
}

func (f *fakeSettingsStore) UpdateTarget(ctx context.Context, target models.Target) (models.Target, error) {
	panic("unused")
}

func (f *fakeSettingsStore) DeleteTarget(ctx context.Context, id int64) error {
	panic("unused")
}

func (f *fakeSettingsStore) ListNotificationsByUser(ctx context.Context, userID int64) ([]models.Notification, error) {
	panic("unused")
}

func (f *fakeSettingsStore) MarkNotificationReadByUser(ctx context.Context, userID, id int64) (models.Notification, error) {
	panic("unused")
}

func (f *fakeSettingsStore) GetUserSettings(ctx context.Context, userID int64) (models.UserSettings, error) {
	f.getCalled = true
	f.getUserID = userID
	return f.getSettings, f.getErr
}

func (f *fakeSettingsStore) UpdateUserSettings(ctx context.Context, userID int64, notificationsEnabled bool) (models.UserSettings, error) {
	f.updateCalled = true
	f.updateUserID = userID
	f.updateValue = notificationsEnabled
	return models.UserSettings{NotificationsEnabled: notificationsEnabled}, f.updateErr
}

func TestHandleSettingsGetReturnsCurrentValue(t *testing.T) {
	store := &fakeSettingsStore{
		getSettings: models.UserSettings{NotificationsEnabled: true},
	}
	server := &APIServer{
		store:  store,
		logger: log.New(io.Discard, "", 0),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/settings", nil)
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.Principal{ID: 7, Email: "user@example.com"}))
	rr := httptest.NewRecorder()

	server.handleSettings(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if !store.getCalled {
		t.Fatal("expected GetUserSettings to be called")
	}
	if store.getUserID != 7 {
		t.Fatalf("unexpected user id: %d", store.getUserID)
	}

	var resp settingsResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if !resp.NotificationsEnabled {
		t.Fatal("expected notifications_enabled=true")
	}
}

func TestHandleSettingsPutUpdatesValue(t *testing.T) {
	store := &fakeSettingsStore{}
	server := &APIServer{
		store:  store,
		logger: log.New(io.Discard, "", 0),
	}

	body := bytes.NewBufferString(`{"notifications_enabled":false}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/settings", body)
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.Principal{ID: 7, Email: "user@example.com"}))
	rr := httptest.NewRecorder()

	server.handleSettings(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if !store.updateCalled {
		t.Fatal("expected UpdateUserSettings to be called")
	}
	if store.updateUserID != 7 {
		t.Fatalf("unexpected user id: %d", store.updateUserID)
	}
	if store.updateValue {
		t.Fatal("expected notifications_enabled=false")
	}

	var resp settingsResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.NotificationsEnabled {
		t.Fatal("expected notifications_enabled=false")
	}
}

func TestHandleSettingsRequiresAuth(t *testing.T) {
	server := &APIServer{
		store:  &fakeSettingsStore{},
		logger: log.New(io.Discard, "", 0),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/settings", nil)
	rr := httptest.NewRecorder()

	server.handleSettings(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
}

func TestHandleSettingsRejectsUnsupportedMethod(t *testing.T) {
	server := &APIServer{
		store:  &fakeSettingsStore{},
		logger: log.New(io.Discard, "", 0),
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/settings", nil)
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.Principal{ID: 7, Email: "user@example.com"}))
	rr := httptest.NewRecorder()

	server.handleSettings(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
}

func TestHandleSettingsReturnsInternalErrorOnStoreFailure(t *testing.T) {
	store := &fakeSettingsStore{
		getErr: errors.New("db failed"),
	}
	server := &APIServer{
		store:  store,
		logger: log.New(io.Discard, "", 0),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/settings", nil)
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.Principal{ID: 7, Email: "user@example.com"}))
	rr := httptest.NewRecorder()

	server.handleSettings(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
}

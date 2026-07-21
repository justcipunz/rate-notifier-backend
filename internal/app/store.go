package app

import (
	"context"
	"time"

	"github.com/justcipunz/rate-notifier-backend/internal/models"
)

type APIStore interface {
	CreateUser(ctx context.Context, email, passwordHash string) (models.User, error)
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
	GetUserByID(ctx context.Context, id int64) (models.User, error)
	ListRates(ctx context.Context) ([]models.Rate, error)
	ListRateHistory(ctx context.Context, currency string, from time.Time, to time.Time) ([]models.RateHistoryPoint, error)
	ListTargetsByUser(ctx context.Context, userID int64) ([]models.Target, error)
	CreateTarget(ctx context.Context, target models.Target) (models.Target, error)
	GetTargetByID(ctx context.Context, id int64) (models.Target, error)
	UpdateTarget(ctx context.Context, target models.Target) (models.Target, error)
	DeleteTarget(ctx context.Context, id int64) error
	ListNotificationsByUser(ctx context.Context, userID int64) ([]models.Notification, error)
	MarkNotificationReadByUser(ctx context.Context, userID, id int64) (models.Notification, error)
	GetUserSettings(ctx context.Context, userID int64) (models.UserSettings, error)
	UpdateUserSettings(ctx context.Context, userID int64, notificationsEnabled bool) (models.UserSettings, error)
}

package models

import "time"

type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type UserSettings struct {
	NotificationsEnabled bool `json:"notifications_enabled"`
}

type Rate struct {
	Currency      string    `json:"currency"`
	Name          string    `json:"name"`
	Value         float64   `json:"value"`
	PreviousValue *float64  `json:"previous_value"`
	Change        *float64  `json:"change"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Target struct {
	ID          int64      `json:"id"`
	UserID      int64      `json:"user_id"`
	Currency    string     `json:"currency"`
	TargetValue float64    `json:"target_value"`
	Condition   string     `json:"condition"`
	IsActive    bool       `json:"is_active"`
	TriggeredAt *time.Time `json:"triggered_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type Notification struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	TargetID    *int64    `json:"target_id,omitempty"`
	Currency    string    `json:"currency"`
	TargetValue float64   `json:"target_value"`
	ActualValue float64   `json:"actual_value"`
	Condition   string    `json:"condition"`
	IsRead      bool      `json:"is_read"`
	CreatedAt   time.Time `json:"created_at"`
}

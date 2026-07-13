package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/justcipunz/rate-notifier-backend/internal/models"
)

func (s *Store) CreateUser(ctx context.Context, email, passwordHash string) (models.User, error) {
	var user models.User

	err := s.pool.QueryRow(ctx, `
INSERT INTO users (email, password_hash)
VALUES ($1, $2)
RETURNING id, email, password_hash, created_at`,
		email, passwordHash,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return models.User{}, ErrEmailExists
		}
		return models.User{}, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	var user models.User

	err := s.pool.QueryRow(ctx, `
SELECT id, email, password_hash, created_at
FROM users
WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, ErrNotFound
		}
		return models.User{}, fmt.Errorf("get user by email: %w", err)
	}

	return user, nil
}

func (s *Store) GetUserByID(ctx context.Context, id int64) (models.User, error) {
	var user models.User

	err := s.pool.QueryRow(ctx, `
SELECT id, email, password_hash, created_at
FROM users
WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, ErrNotFound
		}
		return models.User{}, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}

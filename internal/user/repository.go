package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetByID returns a user from the database given their UUID.
func (r *Repository) GetByID(ctx context.Context, id string) (*User, error) {
	query := `
		SELECT id, email, username, password_hash, avatar_url, role, created_at, updated_at
		FROM users
		WHERE id = ?
	`

	var u User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID,
		&u.Email,
		&u.Username,
		&u.PasswordHash,
		&u.AvatarURL,
		&u.Role,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("database query failed: %w", err)
	}

	return &u, nil
}
func (r *Repository) GetEmailid(ctx context.Context, email string) (*User, error) {
	query := `SELECT id, email, username, password_hash, avatar_url, role, created_at, updated_at
		FROM users
		WHERE email = ?`
	var u User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&u.ID,
		&u.Email,
		&u.Username,
		&u.PasswordHash,
		&u.AvatarURL,
		&u.Role,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("database query failed: %w", err)
	}
	return &u, nil
}
func (r *Repository) CreateUser(ctx context.Context, u *User) error {
	query := `INSERT INTO users (id, email, username, password_hash, avatar_url, role, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	err := r.db.QueryRowContext(ctx, query, u.ID, u.Email, u.Username, u.PasswordHash, u.AvatarURL, u.Role, u.CreatedAt, u.UpdatedAt).Scan(&u.ID)
	if err != nil {
		return fmt.Errorf("database query failed: %w", err)
	}
	return nil
}
func (r *Repository) UpdateUser(ctx context.Context, u *User) error {
	query := `UPDATE users SET email = ?, username = ?, password_hash = ?, avatar_url = ?, role = ?, created_at = ?, updated_at = ? WHERE id = ?`
	err := r.db.QueryRowContext(ctx, query, u.Email, u.Username, u.PasswordHash, u.AvatarURL, u.Role, u.CreatedAt, u.UpdatedAt, u.ID).Scan(&u.ID)
	if err != nil {
		return fmt.Errorf("database query failed: %w", err)
	}
	return nil
}
func (r *Repository) DeleteUser(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE ID = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("database query failed: %w", err)
	}
	return nil
}

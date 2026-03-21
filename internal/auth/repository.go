package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// =============================================================================
// REPOSITORY LAYER (Database access)
// =============================================================================
// The repository is responsible ONLY for talking to the database.
// It executing SQL queries and mapping the results to Go structs.
// It should not contain any business logic (like hashing passwords).

type Repository struct {
	db *sql.DB
}

// NewRepository creates a new instance of the auth repository.
// It takes a connected database connection pool as an argument.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// CreateUser inserts a new user record into the database.
// MySQL's UUID() function is used directly in the SQL statement.
func (r *Repository) CreateUser(ctx context.Context, user *User) error {
	// The id, created_at, role, and updated_at columns have defaults set in SQL schema.
	// We use '?' as placeholders to prevent SQL injection.
	query := `
		INSERT INTO users (email, username, password_hash)
		VALUES (?, ?, ?)
	`

	// ExecContext executes the query without returning any rows.
	_, err := r.db.ExecContext(ctx, query, user.Email, user.Username, user.PasswordHash)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByEmail retrieves a user record from the database by their email.
func (r *Repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, email, username, password_hash, avatar_url, role, created_at, updated_at
		FROM users
		WHERE email = ?
	`

	var user User
	// QueryRowContext expects exactly one row.
	// We use Scan() to copy the result columns into our struct fields.
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.AvatarURL,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// This is a normal scenario, just return our standard error that indicates "not found"
			return nil, fmt.Errorf("user with email %s not found: %w", email, err)
		}
		// Some other database error occurred
		return nil, fmt.Errorf("failed to fetch user by email: %w", err)
	}

	return &user, nil
}

// EmailExists checks if an email is already registered in the database.
// This is used during registration to prevent duplicate accounts.
func (r *Repository) EmailExists(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return exists, nil
}

// UsernameExists checks if a username is already taken.
func (r *Repository) UsernameExists(ctx context.Context, username string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check username existence: %w", err)
	}

	return exists, nil
}

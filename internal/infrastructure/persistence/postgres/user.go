package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/dmitokk/FragmentsBE/internal/domain/entity"
	"github.com/dmitokk/FragmentsBE/internal/domain/repository"
	"github.com/google/uuid"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) repository.UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, google_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.GoogleID,
		time.Now(),
		time.Now(),
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
		SELECT id, email, password_hash, google_id, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user entity.User
	var googleID sql.NullString

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&googleID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if googleID.Valid {
		user.GoogleID = googleID.String
	}

	return &user, nil
}

func (r *UserRepository) GetByGoogleID(ctx context.Context, googleID string) (*entity.User, error) {
	query := `
		SELECT id, email, password_hash, google_id, created_at, updated_at
		FROM users
		WHERE google_id = $1
	`

	var user entity.User
	var googleIDNull sql.NullString

	err := r.db.QueryRowContext(ctx, query, googleID).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&googleIDNull,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if googleIDNull.Valid {
		user.GoogleID = googleIDNull.String
	}

	return &user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	query := `
		SELECT id, email, password_hash, google_id, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user entity.User
	var googleID sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&googleID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if googleID.Valid {
		user.GoogleID = googleID.String
	}

	return &user, nil
}
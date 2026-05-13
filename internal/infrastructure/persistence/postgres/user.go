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
		INSERT INTO users (id, email, password_hash, google_id, name, avatar_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.GoogleID,
		nullIfEmpty(user.Name),
		nullIfEmpty(user.AvatarURL),
		time.Now(),
		time.Now(),
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
		SELECT id, email, password_hash, google_id, name, avatar_url, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user entity.User
	var googleID, name, avatarURL sql.NullString

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&googleID,
		&name,
		&avatarURL,
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
	if name.Valid {
		user.Name = name.String
	}
	if avatarURL.Valid {
		user.AvatarURL = avatarURL.String
	}

	return &user, nil
}

func (r *UserRepository) GetByGoogleID(ctx context.Context, googleID string) (*entity.User, error) {
	query := `
		SELECT id, email, password_hash, google_id, name, avatar_url, created_at, updated_at
		FROM users
		WHERE google_id = $1
	`

	var user entity.User
	var googleIDNull, name, avatarURL sql.NullString

	err := r.db.QueryRowContext(ctx, query, googleID).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&googleIDNull,
		&name,
		&avatarURL,
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
	if name.Valid {
		user.Name = name.String
	}
	if avatarURL.Valid {
		user.AvatarURL = avatarURL.String
	}

	return &user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	query := `
		SELECT id, email, password_hash, google_id, name, avatar_url, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user entity.User
	var googleID, name, avatarURL sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&googleID,
		&name,
		&avatarURL,
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
	if name.Valid {
		user.Name = name.String
	}
	if avatarURL.Valid {
		user.AvatarURL = avatarURL.String
	}

	return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *entity.User) error {
	query := `
		UPDATE users
		SET name = $2, avatar_url = $3, updated_at = $4
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		user.ID,
		nullIfEmpty(user.Name),
		nullIfEmpty(user.AvatarURL),
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}
package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dmitokk/FragmentsBE/internal/domain/entity"
	"github.com/google/uuid"
)

type UserFragmentRepository struct {
	db *sql.DB
}

func NewUserFragmentRepository(db *sql.DB) *UserFragmentRepository {
	return &UserFragmentRepository{db: db}
}

func (r *UserFragmentRepository) Create(ctx context.Context, userID, fragmentID uuid.UUID) error {
	query := `
		INSERT INTO user_fragments (user_id, fragment_id, found_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, userID, fragmentID)
	if err != nil {
		return fmt.Errorf("failed to create user fragment: %w", err)
	}

	return nil
}

func (r *UserFragmentRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.UserFragment, error) {
	query := `
		SELECT user_id, fragment_id, found_at
		FROM user_fragments
		WHERE user_id = $1
		ORDER BY found_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user fragments: %w", err)
	}
	defer rows.Close()

	var fragments []*entity.UserFragment
	for rows.Next() {
		var uf entity.UserFragment
		err := rows.Scan(&uf.UserID, &uf.FragmentID, &uf.FoundAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user fragment: %w", err)
		}
		fragments = append(fragments, &uf)
	}

	return fragments, nil
}

func (r *UserFragmentRepository) Exists(ctx context.Context, userID, fragmentID uuid.UUID) (bool, error) {
	query := `
		SELECT 1 FROM user_fragments
		WHERE user_id = $1 AND fragment_id = $2
	`

	var exists int
	err := r.db.QueryRowContext(ctx, query, userID, fragmentID).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check user fragment: %w", err)
	}

	return true, nil
}

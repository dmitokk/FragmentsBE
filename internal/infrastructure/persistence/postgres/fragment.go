package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/dmitokk/FragmentsBE/internal/domain/entity"
	"github.com/dmitokk/FragmentsBE/internal/domain/repository"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type FragmentRepository struct {
	db *sql.DB
}

func NewFragmentRepository(db *sql.DB) repository.FragmentRepository {
	return &FragmentRepository{db: db}
}

func (r *FragmentRepository) Create(ctx context.Context, fragment *entity.Fragment) error {
	query := `
		INSERT INTO fragments (id, user_id, text, geomark, sound_url, photo_urls, created_at, updated_at)
		VALUES ($1, $2, $3, ST_SetSRID(ST_MakePoint($4, $5), 4326), $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		fragment.ID,
		fragment.UserID,
		fragment.Text,
		fragment.Geomark.Lng,
		fragment.Geomark.Lat,
		fragment.SoundURL,
		pq.Array(fragment.PhotoURLs),
		time.Now(),
		time.Now(),
	).Scan(&fragment.ID, &fragment.CreatedAt, &fragment.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create fragment: %w", err)
	}

	return nil
}

func (r *FragmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Fragment, error) {
	query := `
		SELECT id, user_id, text, ST_X(geomark) as lng, ST_Y(geomark) as lat, 
		       sound_url, photo_urls, created_at, updated_at
		FROM fragments
		WHERE id = $1
	`

	var fragment entity.Fragment
	var lat, lng sql.NullFloat64
	var photoUrls pq.StringArray

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&fragment.ID,
		&fragment.UserID,
		&fragment.Text,
		&lng,
		&lat,
		&fragment.SoundURL,
		&photoUrls,
		&fragment.CreatedAt,
		&fragment.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("fragment not found")
		}
		return nil, fmt.Errorf("failed to get fragment: %w", err)
	}

	if lat.Valid && lng.Valid {
		fragment.Geomark = &entity.Geomark{Lat: lat.Float64, Lng: lng.Float64}
	}

	fragment.PhotoURLs = photoUrls

	return &fragment, nil
}

func (r *FragmentRepository) List(ctx context.Context, userID uuid.UUID, lat, lng, radius float64) ([]*entity.Fragment, error) {
	query := `
		SELECT id, user_id, text, ST_X(geomark) as lng, ST_Y(geomark) as lat, 
		       sound_url, photo_urls, created_at, updated_at
		FROM fragments
		WHERE user_id = $1
		AND ST_DWithin(geomark::geography, ST_MakePoint($2, $3)::geography, $4)
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, lng, lat, radius)
	if err != nil {
		return nil, fmt.Errorf("failed to list fragments: %w", err)
	}
	defer rows.Close()

	var fragments []*entity.Fragment
	for rows.Next() {
		var fragment entity.Fragment
		var lat, lng sql.NullFloat64
		var photoUrls pq.StringArray

		err := rows.Scan(
			&fragment.ID,
			&fragment.UserID,
			&fragment.Text,
			&lng,
			&lat,
			&fragment.SoundURL,
			&photoUrls,
			&fragment.CreatedAt,
			&fragment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan fragment: %w", err)
		}

		if lat.Valid && lng.Valid {
			fragment.Geomark = &entity.Geomark{Lat: lat.Float64, Lng: lng.Float64}
		}

		fragment.PhotoURLs = photoUrls
		fragments = append(fragments, &fragment)
	}

	return fragments, nil
}

func (r *FragmentRepository) Update(ctx context.Context, fragment *entity.Fragment) error {
	query := `
		UPDATE fragments
		SET text = $2, geomark = ST_SetSRID(ST_MakePoint($3, $4), 4326), 
		    sound_url = $5, photo_urls = $6, updated_at = $7
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		fragment.ID,
		fragment.Text,
		fragment.Geomark.Lng,
		fragment.Geomark.Lat,
		fragment.SoundURL,
		pq.Array(fragment.PhotoURLs),
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to update fragment: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("fragment not found")
	}

	return nil
}

func (r *FragmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM fragments WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete fragment: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("fragment not found")
	}

	return nil
}
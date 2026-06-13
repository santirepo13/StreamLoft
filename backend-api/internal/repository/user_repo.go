package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"streamloft-api/internal/models"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByNumericID(ctx context.Context, numericID string) (*models.User, error) {
	query := `
		SELECT id, numeric_id, name, stream_key, bitrate, created_at, updated_at
		FROM users
		WHERE numeric_id = $1
	`

	var user models.User
	err := r.db.QueryRow(ctx, query, numericID).Scan(
		&user.ID,
		&user.NumericID,
		&user.Name,
		&user.StreamKey,
		&user.Bitrate,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int) (*models.User, error) {
	query := `
		SELECT id, numeric_id, name, stream_key, bitrate, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user models.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.NumericID,
		&user.Name,
		&user.StreamKey,
		&user.Bitrate,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) Create(ctx context.Context, numericID, name, streamKey string) (*models.User, error) {
	query := `
		INSERT INTO users (numeric_id, name, stream_key)
		VALUES ($1, $2, $3)
		RETURNING id, numeric_id, name, stream_key, bitrate, created_at, updated_at
	`

	var user models.User
	err := r.db.QueryRow(ctx, query, numericID, name, streamKey).Scan(
		&user.ID,
		&user.NumericID,
		&user.Name,
		&user.StreamKey,
		&user.Bitrate,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	log.Info().Str("numeric_id", numericID).Int("user_id", user.ID).Msg("user created")
	return &user, nil
}

func (r *UserRepository) UpdateStreamKey(ctx context.Context, userID int, streamKey string) error {
	query := `
		UPDATE users
		SET stream_key = $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := r.db.Exec(ctx, query, streamKey, userID)
	if err != nil {
		return fmt.Errorf("failed to update stream key: %w", err)
	}

	return nil
}

func (r *UserRepository) GetDestinationsByUserID(ctx context.Context, userID int) ([]models.UserDestination, error) {
	query := `
		SELECT ud.id, ud.user_id, ud.destination_id, ud.stream_key, ud.enabled,
		       ud.created_at, ud.updated_at,
		       d.id, d.name, d.rtmp_url, d.created_at
		FROM user_destinations ud
		JOIN destinations d ON ud.destination_id = d.id
		WHERE ud.user_id = $1
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get destinations: %w", err)
	}
	defer rows.Close()

	var destinations []models.UserDestination
	for rows.Next() {
		var ud models.UserDestination
		var d models.Destination
		err := rows.Scan(
			&ud.ID, &ud.UserID, &ud.DestinationID, &ud.StreamKey, &ud.Enabled,
			&ud.CreatedAt, &ud.UpdatedAt,
			&d.ID, &d.Name, &d.RTMPURL, &d.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan destination: %w", err)
		}
		destination := d
		ud.Destination = &destination
		destinations = append(destinations, ud)
	}

	return destinations, nil
}

func (r *UserRepository) GetByStreamKey(ctx context.Context, streamKey string) (*models.User, error) {
	query := `
		SELECT id, numeric_id, name, stream_key, bitrate, created_at, updated_at
		FROM users
		WHERE stream_key = $1
	`

	var user models.User
	err := r.db.QueryRow(ctx, query, streamKey).Scan(
		&user.ID,
		&user.NumericID,
		&user.Name,
		&user.StreamKey,
		&user.Bitrate,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by stream key: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) UpdateBitrate(ctx context.Context, userID int, bitrate int) error {
	query := `
		UPDATE users
		SET bitrate = $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := r.db.Exec(ctx, query, bitrate, userID)
	if err != nil {
		return fmt.Errorf("failed to update bitrate: %w", err)
	}

	return nil
}

func (r *UserRepository) GetDestinationsWithStreamKey(ctx context.Context, userID int) ([]models.UserDestination, error) {
	query := `
		SELECT ud.id, ud.user_id, ud.destination_id, ud.stream_key, ud.enabled,
		       ud.created_at, ud.updated_at,
		       d.id, d.name, d.rtmp_url, d.created_at
		FROM user_destinations ud
		JOIN destinations d ON ud.destination_id = d.id
		WHERE ud.user_id = $1 AND ud.stream_key IS NOT NULL AND ud.stream_key != '' AND ud.enabled = 1
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get destinations with stream key: %w", err)
	}
	defer rows.Close()

	var destinations []models.UserDestination
	for rows.Next() {
		var ud models.UserDestination
		var d models.Destination
		err := rows.Scan(
			&ud.ID, &ud.UserID, &ud.DestinationID, &ud.StreamKey, &ud.Enabled,
			&ud.CreatedAt, &ud.UpdatedAt,
			&d.ID, &d.Name, &d.RTMPURL, &d.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan destination: %w", err)
		}
		destination := d
		ud.Destination = &destination
		destinations = append(destinations, ud)
	}

	return destinations, nil
}

func (r *UserRepository) GetDestinationByID(ctx context.Context, userID, destinationID int) (*models.UserDestination, error) {
	query := `
		SELECT ud.id, ud.user_id, ud.destination_id, ud.stream_key, ud.enabled,
		       ud.created_at, ud.updated_at,
		       d.id, d.name, d.rtmp_url, d.created_at
		FROM user_destinations ud
		JOIN destinations d ON ud.destination_id = d.id
		WHERE ud.user_id = $1 AND ud.id = $2
	`

	var ud models.UserDestination
	var d models.Destination
	err := r.db.QueryRow(ctx, query, userID, destinationID).Scan(
		&ud.ID, &ud.UserID, &ud.DestinationID, &ud.StreamKey, &ud.Enabled,
		&ud.CreatedAt, &ud.UpdatedAt,
		&d.ID, &d.Name, &d.RTMPURL, &d.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get destination: %w", err)
	}

	ud.Destination = &d
	return &ud, nil
}

func (r *UserRepository) UpdateDestinationStreamKey(ctx context.Context, userID, destinationID int, streamKey string) error {
	query := `
		UPDATE user_destinations
		SET stream_key = $1, updated_at = NOW()
		WHERE user_id = $2 AND id = $3
	`

	_, err := r.db.Exec(ctx, query, streamKey, userID, destinationID)
	if err != nil {
		return fmt.Errorf("failed to update destination stream key: %w", err)
	}

	return nil
}

func (r *UserRepository) UpdateDestinationEnabled(ctx context.Context, userID, destinationID int, enabled int) error {
	query := `
		UPDATE user_destinations
		SET enabled = $1, updated_at = NOW()
		WHERE user_id = $2 AND id = $3
	`

	_, err := r.db.Exec(ctx, query, enabled, userID, destinationID)
	if err != nil {
		return fmt.Errorf("failed to update destination enabled: %w", err)
	}

	return nil
}

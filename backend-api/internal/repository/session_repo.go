package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"streamloft-api/internal/models"
)

type SessionRepository struct {
	db *pgxpool.Pool
}

func NewSessionRepository(db *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, userID int, machineID, accessToken, refreshToken string, expiresAt time.Time) (*models.UserSession, error) {
	query := `
		INSERT INTO user_sessions (user_id, machine_id, access_token, refresh_token, token_expires_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, machine_id) 
		DO UPDATE SET access_token = EXCLUDED.access_token, 
					refresh_token = EXCLUDED.refresh_token,
					token_expires_at = EXCLUDED.token_expires_at,
					updated_at = NOW()
		RETURNING id, user_id, machine_id, token_expires_at, created_at, updated_at
	`

	var session models.UserSession
	err := r.db.QueryRow(ctx, query, userID, machineID, accessToken, refreshToken, expiresAt).Scan(
		&session.ID,
		&session.UserID,
		&session.MachineID,
		&session.TokenExpiresAt,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	log.Info().Int("user_id", userID).Str("machine_id", machineID).Msg("session upserted")
	return &session, nil
}

func (r *SessionRepository) GetByRefreshToken(ctx context.Context, refreshToken string) (*models.UserSession, error) {
	query := `
		SELECT id, user_id, machine_id, access_token, refresh_token, token_expires_at, created_at, updated_at
		FROM user_sessions
		WHERE refresh_token = $1 AND token_expires_at > NOW()
	`

	var session models.UserSession
	err := r.db.QueryRow(ctx, query, refreshToken).Scan(
		&session.ID,
		&session.UserID,
		&session.MachineID,
		&session.AccessToken,
		&session.RefreshToken,
		&session.TokenExpiresAt,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session by refresh token: %w", err)
	}

	return &session, nil
}

func (r *SessionRepository) GetByAccessToken(ctx context.Context, accessToken string) (*models.UserSession, error) {
	query := `
		SELECT id, user_id, machine_id, access_token, refresh_token, token_expires_at, created_at, updated_at
		FROM user_sessions
		WHERE access_token = $1 AND token_expires_at > NOW()
	`

	var session models.UserSession
	err := r.db.QueryRow(ctx, query, accessToken).Scan(
		&session.ID,
		&session.UserID,
		&session.MachineID,
		&session.AccessToken,
		&session.RefreshToken,
		&session.TokenExpiresAt,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session by access token: %w", err)
	}

	return &session, nil
}

func (r *SessionRepository) DeleteByUserAndMachine(ctx context.Context, userID int, machineID string) error {
	query := `
		DELETE FROM user_sessions
		WHERE user_id = $1 AND machine_id = $2
	`

	result, err := r.db.Exec(ctx, query, userID, machineID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	if result.RowsAffected() > 0 {
		log.Info().Int("user_id", userID).Str("machine_id", machineID).Msg("session deleted")
	}

	return nil
}

func (r *SessionRepository) UpdateAccessToken(ctx context.Context, sessionID int, accessToken string, expiresAt time.Time) error {
	query := `
		UPDATE user_sessions
		SET access_token = $1, token_expires_at = $2, updated_at = NOW()
		WHERE id = $3
	`

	_, err := r.db.Exec(ctx, query, accessToken, expiresAt, sessionID)
	if err != nil {
		return fmt.Errorf("failed to update access token: %w", err)
	}

	return nil
}
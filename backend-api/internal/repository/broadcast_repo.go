package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"streamloft-api/internal/interfaces"
	"streamloft-api/internal/models"
)

type BroadcastRepository struct {
	db *pgxpool.Pool
}

func NewBroadcastRepository(db *pgxpool.Pool) *BroadcastRepository {
	return &BroadcastRepository{db: db}
}

func (r *BroadcastRepository) Create(ctx context.Context, userID, userDestinationID int) (*models.BroadcastSession, error) {
	query := `
		INSERT INTO broadcast_sessions (user_id, user_destination_id, date, started_at)
		VALUES ($1, $2, CURRENT_DATE, NOW())
		RETURNING id, user_id, user_destination_id, date, duration_minutes, started_at, ended_at
	`

	var session models.BroadcastSession
	err := r.db.QueryRow(ctx, query, userID, userDestinationID).Scan(
		&session.ID,
		&session.UserID,
		&session.UserDestinationID,
		&session.Date,
		&session.DurationMinutes,
		&session.StartedAt,
		&session.EndedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create broadcast session: %w", err)
	}

	log.Info().Int("user_id", userID).Int("destination_id", userDestinationID).Msg("broadcast session created")
	return &session, nil
}

func (r *BroadcastRepository) UpdateEndTime(ctx context.Context, sessionID int, endedAt time.Time, durationMinutes int) error {
	query := `
		UPDATE broadcast_sessions
		SET ended_at = $1, duration_minutes = $2
		WHERE id = $3
	`

	_, err := r.db.Exec(ctx, query, endedAt, durationMinutes, sessionID)
	if err != nil {
		return fmt.Errorf("failed to update broadcast session: %w", err)
	}

	return nil
}

func (r *BroadcastRepository) GetActiveByUserID(ctx context.Context, userID int) ([]models.BroadcastSession, error) {
	query := `
		SELECT id, user_id, user_destination_id, date, duration_minutes, started_at, ended_at
		FROM broadcast_sessions
		WHERE user_id = $1 AND ended_at IS NULL
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active broadcast sessions: %w", err)
	}
	defer rows.Close()

	var sessions []models.BroadcastSession
	for rows.Next() {
		var session models.BroadcastSession
		err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.UserDestinationID,
			&session.Date,
			&session.DurationMinutes,
			&session.StartedAt,
			&session.EndedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan broadcast session: %w", err)
		}
		sessions = append(sessions, session)
	}

	if sessions == nil {
		return []models.BroadcastSession{}, nil
	}

	return sessions, nil
}

func (r *BroadcastRepository) GetByUserID(ctx context.Context, userID int) ([]models.BroadcastSession, error) {
	query := `
		SELECT id, user_id, user_destination_id, date, duration_minutes, started_at, ended_at
		FROM broadcast_sessions
		WHERE user_id = $1
		ORDER BY started_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get broadcast sessions: %w", err)
	}
	defer rows.Close()

	var sessions []models.BroadcastSession
	for rows.Next() {
		var session models.BroadcastSession
		err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.UserDestinationID,
			&session.Date,
			&session.DurationMinutes,
			&session.StartedAt,
			&session.EndedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan broadcast session: %w", err)
		}
		sessions = append(sessions, session)
	}

	if sessions == nil {
		return []models.BroadcastSession{}, nil
	}

	return sessions, nil
}
func (r *BroadcastRepository) DeleteByDestinationAndDate(ctx context.Context, userID int, destinationName string, date time.Time) (int64, error) {
	query := `
		DELETE FROM broadcast_sessions
		WHERE user_id = $1
		  AND user_destination_id IN (
		      SELECT ud.id
		      FROM user_destinations ud
		      JOIN destinations d ON ud.destination_id = d.id
		      WHERE d.name = $2
		  )
		  AND date = $3
	`

	tag, err := r.db.Exec(ctx, query, userID, destinationName, date)
	if err != nil {
		return 0, fmt.Errorf("failed to delete broadcast sessions: %w", err)
	}

	deleted := tag.RowsAffected()
	log.Info().
		Int("user_id", userID).
		Str("destination_name", destinationName).
		Str("date", date.Format("2006-01-02")).
		Int64("deleted", deleted).
		Msg("broadcast sessions deleted by destination and date")
	return deleted, nil
}

func (r *BroadcastRepository) GetByUserIDWithDestination(ctx context.Context, userID int) ([]interfaces.BroadcastSessionWithDestination, error) {
	query := `
		SELECT bs.id, bs.user_id, bs.user_destination_id, d.name,
		       bs.date, bs.duration_minutes, bs.started_at, bs.ended_at
		FROM broadcast_sessions bs
		JOIN user_destinations ud ON bs.user_destination_id = ud.id
		JOIN destinations d ON ud.destination_id = d.id
		WHERE bs.user_id = $1
		ORDER BY bs.started_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get broadcast sessions with destination: %w", err)
	}
	defer rows.Close()

	var sessions []interfaces.BroadcastSessionWithDestination
	for rows.Next() {
		var session interfaces.BroadcastSessionWithDestination
		err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.UserDestinationID,
			&session.DestinationName,
			&session.Date,
			&session.DurationMinutes,
			&session.StartedAt,
			&session.EndedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan broadcast session: %w", err)
		}
		sessions = append(sessions, session)
	}

	if sessions == nil {
		return []interfaces.BroadcastSessionWithDestination{}, nil
	}

	return sessions, nil
}

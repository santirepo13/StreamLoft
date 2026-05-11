package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"streamloft-api/internal/models"
)

type MachineRepository struct {
	db *pgxpool.Pool
}

func NewMachineRepository(db *pgxpool.Pool) *MachineRepository {
	return &MachineRepository{db: db}
}

func (r *MachineRepository) Upsert(ctx context.Context, machineID string, userID int) (*models.UserMachine, error) {
	query := `
		INSERT INTO user_machines (machine_id, user_id, last_used_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (machine_id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			last_used_at = EXCLUDED.last_used_at
		RETURNING id, machine_id, user_id, last_used_at
	`

	var machine models.UserMachine
	err := r.db.QueryRow(ctx, query, machineID, userID).Scan(
		&machine.ID,
		&machine.MachineID,
		&machine.UserID,
		&machine.LastUsedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert machine: %w", err)
	}

	log.Info().Str("machine_id", machineID).Int("user_id", userID).Msg("machine upserted")
	return &machine, nil
}

func (r *MachineRepository) GetByMachineID(ctx context.Context, machineID string) (*models.UserMachine, error) {
	query := `
		SELECT id, machine_id, user_id, last_used_at
		FROM user_machines
		WHERE machine_id = $1
		ORDER BY last_used_at DESC
		LIMIT 1
	`

	var machine models.UserMachine
	err := r.db.QueryRow(ctx, query, machineID).Scan(
		&machine.ID,
		&machine.MachineID,
		&machine.UserID,
		&machine.LastUsedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get machine: %w", err)
	}

	return &machine, nil
}
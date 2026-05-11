package interfaces

import (
	"context"

	"streamloft-api/internal/models"
)

type MachineRepository interface {
	Upsert(ctx context.Context, machineID string, userID int) (*models.UserMachine, error)
	GetByMachineID(ctx context.Context, machineID string) (*models.UserMachine, error)
}
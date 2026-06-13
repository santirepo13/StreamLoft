package interfaces

import (
	"context"
	"time"

	"streamloft-api/internal/models"
)

type SessionRepository interface {
	Create(ctx context.Context, userID int, machineID, accessToken, refreshToken string, expiresAt time.Time) (*models.UserSession, error)
	GetByRefreshToken(ctx context.Context, refreshToken string) (*models.UserSession, error)
	GetByAccessToken(ctx context.Context, accessToken string) (*models.UserSession, error)
	DeleteByUserAndMachine(ctx context.Context, userID int, machineID string) error
	UpdateAccessToken(ctx context.Context, sessionID int, accessToken string, expiresAt time.Time) error
}
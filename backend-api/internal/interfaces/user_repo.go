package interfaces

import (
	"context"

	"streamloft-api/internal/models"
)

type UserRepository interface {
	GetByID(ctx context.Context, id int) (*models.User, error)
	GetByNumericID(ctx context.Context, numericID string) (*models.User, error)
	GetByStreamKey(ctx context.Context, streamKey string) (*models.User, error)
	Create(ctx context.Context, numericID, name, streamKey string) (*models.User, error)
	UpdateStreamKey(ctx context.Context, userID int, streamKey string) error
	UpdateBitrate(ctx context.Context, userID int, bitrate int) error
	GetDestinationsByUserID(ctx context.Context, userID int) ([]models.UserDestination, error)
	GetDestinationsWithStreamKey(ctx context.Context, userID int) ([]models.UserDestination, error)
	GetDestinationByID(ctx context.Context, userID, destinationID int) (*models.UserDestination, error)
	UpdateDestinationStreamKey(ctx context.Context, userID, destinationID int, streamKey string) error
}
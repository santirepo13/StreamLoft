package interfaces

import "context"

type WorkerManager interface {
	StartWorker(ctx context.Context, userID int, userDestinationID int) error
	StopWorker(ctx context.Context, userID int, userDestinationID int) error
	StopAllForUser(ctx context.Context, userID int) error
	IsWorkerRunning(userID, userDestinationID int) bool
}
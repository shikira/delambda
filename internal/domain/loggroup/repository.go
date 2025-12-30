package loggroup

import "context"

// Repository defines the interface for LogGroup persistence
type Repository interface {
	// Exists checks if a log group exists
	Exists(ctx context.Context, logGroup *LogGroup) (bool, error)

	// Delete deletes a log group
	Delete(ctx context.Context, logGroup *LogGroup) error
}

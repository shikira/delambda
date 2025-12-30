package function

import "context"

// Repository defines the interface for Lambda function persistence
type Repository interface {
	// FindAll returns all Lambda functions
	FindAll(ctx context.Context) ([]*Function, error)

	// FindByName finds a Lambda function by name
	FindByName(ctx context.Context, name string) (*Function, error)

	// DisableIPv6 disables IPv6 for a Lambda function
	DisableIPv6(ctx context.Context, functionName string) error

	// DetachVPC detaches VPC from a Lambda function
	DetachVPC(ctx context.Context, functionName string) error

	// Delete deletes a Lambda function
	Delete(ctx context.Context, functionName string) error
}

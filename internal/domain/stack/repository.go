package stack

import "context"

// Repository defines the interface for stack operations
type Repository interface {
	// ListLambdaFunctions returns all Lambda function names in the specified stack
	ListLambdaFunctions(ctx context.Context, stackName string) ([]string, error)
}

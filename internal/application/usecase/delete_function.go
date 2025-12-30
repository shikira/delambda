package usecase

import (
	"context"
	"fmt"

	"github.com/shirasu/clambda/internal/domain/function"
	"github.com/shirasu/clambda/internal/domain/loggroup"
)

// DeleteFunctionUseCase handles the deletion of Lambda functions
type DeleteFunctionUseCase struct {
	functionRepo function.Repository
	logGroupRepo loggroup.Repository
}

// DeleteFunctionInput represents the input for deleting a function
type DeleteFunctionInput struct {
	FunctionName string
	DetachVPC    bool
	DisableIPv6  bool
	DeleteLogs   bool
}

// NewDeleteFunctionUseCase creates a new DeleteFunctionUseCase
func NewDeleteFunctionUseCase(
	functionRepo function.Repository,
	logGroupRepo loggroup.Repository,
) *DeleteFunctionUseCase {
	return &DeleteFunctionUseCase{
		functionRepo: functionRepo,
		logGroupRepo: logGroupRepo,
	}
}

// Execute executes the delete function use case
func (uc *DeleteFunctionUseCase) Execute(ctx context.Context, input *DeleteFunctionInput) error {
	// Detach VPC if requested
	if input.DetachVPC {
		// Disable IPv6 if requested
		if input.DisableIPv6 {
			if err := uc.functionRepo.DisableIPv6(ctx, input.FunctionName); err != nil {
				// Continue if the function is not attached to a VPC
				if err.Error() != fmt.Sprintf("function %s is not attached to a VPC", input.FunctionName) {
					return fmt.Errorf("failed to disable IPv6: %w", err)
				}
			}
		}

		// Detach VPC
		if err := uc.functionRepo.DetachVPC(ctx, input.FunctionName); err != nil {
			// Continue if the function is not attached to a VPC
			if err.Error() != fmt.Sprintf("function %s is not attached to a VPC", input.FunctionName) {
				return fmt.Errorf("failed to detach VPC: %w", err)
			}
		}
	}

	// Delete the function
	if err := uc.functionRepo.Delete(ctx, input.FunctionName); err != nil {
		return fmt.Errorf("failed to delete function: %w", err)
	}

	// Delete log group if requested
	if input.DeleteLogs {
		logGroup := loggroup.NewLogGroupForFunction(input.FunctionName)
		if err := uc.logGroupRepo.Delete(ctx, logGroup); err != nil {
			return fmt.Errorf("failed to delete log group: %w", err)
		}
	}

	return nil
}

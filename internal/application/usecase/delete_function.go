package usecase

import (
	"context"
	"fmt"
	"io"

	"github.com/shirasu/delambda/internal/domain/function"
	"github.com/shirasu/delambda/internal/domain/loggroup"
)

// DeleteFunctionUseCase handles the deletion of Lambda functions
type DeleteFunctionUseCase struct {
	functionRepo function.Repository
	logGroupRepo loggroup.Repository
	output       io.Writer
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
	output io.Writer,
) *DeleteFunctionUseCase {
	return &DeleteFunctionUseCase{
		functionRepo: functionRepo,
		logGroupRepo: logGroupRepo,
		output:       output,
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
				if err.Error() == fmt.Sprintf("function %s is not attached to a VPC", input.FunctionName) {
					fmt.Fprintf(uc.output, "Function is not attached to VPC, skipping IPv6 disable\n")
				} else {
					return fmt.Errorf("failed to disable IPv6: %w", err)
				}
			} else {
				fmt.Fprintf(uc.output, "Disabled IPv6 for function %s\n", input.FunctionName)
			}
		}

		// Detach VPC
		if err := uc.functionRepo.DetachVPC(ctx, input.FunctionName); err != nil {
			// Continue if the function is not attached to a VPC
			if err.Error() == fmt.Sprintf("function %s is not attached to a VPC", input.FunctionName) {
				fmt.Fprintf(uc.output, "Function is not attached to VPC, skipping VPC detach\n")
			} else {
				return fmt.Errorf("failed to detach VPC: %w", err)
			}
		} else {
			fmt.Fprintf(uc.output, "Detached VPC from function %s\n", input.FunctionName)
		}
	}

	// Delete the function
	fmt.Fprintf(uc.output, "Deleting function %s...\n", input.FunctionName)
	if err := uc.functionRepo.Delete(ctx, input.FunctionName); err != nil {
		return fmt.Errorf("failed to delete function: %w", err)
	}
	fmt.Fprintf(uc.output, "Deleted function %s\n", input.FunctionName)

	// Delete log group if requested
	if input.DeleteLogs {
		logGroup := loggroup.NewLogGroupForFunction(input.FunctionName)
		fmt.Fprintf(uc.output, "Deleting CloudWatch Logs log group %s...\n", logGroup.Name())
		if err := uc.logGroupRepo.Delete(ctx, logGroup); err != nil {
			return fmt.Errorf("failed to delete log group: %w", err)
		}
		fmt.Fprintf(uc.output, "Deleted CloudWatch Logs log group %s\n", logGroup.Name())
	} else {
		fmt.Fprintf(uc.output, "Skipping log deletion (--without-logs specified)\n")
	}

	return nil
}

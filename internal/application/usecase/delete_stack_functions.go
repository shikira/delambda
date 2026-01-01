package usecase

import (
	"context"
	"fmt"
	"io"

	"github.com/shirasu/delambda/internal/domain/function"
	"github.com/shirasu/delambda/internal/domain/loggroup"
	"github.com/shirasu/delambda/internal/domain/stack"
)

// DeleteStackFunctionsUseCase handles deleting all Lambda functions in a stack
type DeleteStackFunctionsUseCase struct {
	functionRepo function.Repository
	logGroupRepo loggroup.Repository
	stackRepo    stack.Repository
	output       io.Writer
}

// DeleteStackFunctionsInput contains the input parameters for deleting stack functions
type DeleteStackFunctionsInput struct {
	StackName   string
	DetachVPC   bool
	DisableIPv6 bool
	DeleteLogs  bool
}

// NewDeleteStackFunctionsUseCase creates a new DeleteStackFunctionsUseCase
func NewDeleteStackFunctionsUseCase(
	functionRepo function.Repository,
	logGroupRepo loggroup.Repository,
	stackRepo stack.Repository,
	output io.Writer,
) *DeleteStackFunctionsUseCase {
	return &DeleteStackFunctionsUseCase{
		functionRepo: functionRepo,
		logGroupRepo: logGroupRepo,
		stackRepo:    stackRepo,
		output:       output,
	}
}

// Execute deletes all Lambda functions in the specified stack
func (uc *DeleteStackFunctionsUseCase) Execute(ctx context.Context, input *DeleteStackFunctionsInput) error {
	// Get all Lambda functions in the stack
	functionNames, err := uc.stackRepo.ListLambdaFunctions(ctx, input.StackName)
	if err != nil {
		return fmt.Errorf("failed to list Lambda functions in stack: %w", err)
	}

	if len(functionNames) == 0 {
		return fmt.Errorf("no Lambda functions found in stack %s", input.StackName)
	}

	fmt.Fprintf(uc.output, "Found %d Lambda function(s) in stack %s\n", len(functionNames), input.StackName)

	// Delete each function
	successCount := 0
	failureCount := 0
	for _, functionName := range functionNames {
		fmt.Fprintf(uc.output, "\n=== Processing function: %s ===\n", functionName)

		// Get the function to check VPC status
		fn, err := uc.functionRepo.FindByName(ctx, functionName)
		if err != nil {
			fmt.Fprintf(uc.output, "Failed to get function: %v\n", err)
			failureCount++
			continue
		}

		// Handle VPC detachment if requested
		if input.DetachVPC {
			if fn.IsAttachedToVPC() {
				// Disable IPv6 if requested and enabled
				if input.DisableIPv6 {
					if fn.HasIPv6Enabled() {
						if err := uc.functionRepo.DisableIPv6(ctx, functionName); err != nil {
							fmt.Fprintf(uc.output, "Failed to disable IPv6: %v\n", err)
							failureCount++
							continue
						}
						fmt.Fprintf(uc.output, "Disabled IPv6 for function %s\n", functionName)
					} else {
						fmt.Fprintf(uc.output, "IPv6 is not enabled, skipping IPv6 disable\n")
					}
				}

				// Detach VPC
				if err := uc.functionRepo.DetachVPC(ctx, functionName); err != nil {
					fmt.Fprintf(uc.output, "Failed to detach VPC: %v\n", err)
					failureCount++
					continue
				}
				fmt.Fprintf(uc.output, "Detached VPC from function %s\n", functionName)
			} else {
				fmt.Fprintf(uc.output, "Function is not attached to VPC, skipping VPC detach\n")
			}
		}

		// Delete the function
		fmt.Fprintf(uc.output, "Deleting function %s...\n", functionName)
		if err := uc.functionRepo.Delete(ctx, functionName); err != nil {
			fmt.Fprintf(uc.output, "Failed to delete function: %v\n", err)
			failureCount++
			continue
		}
		fmt.Fprintf(uc.output, "Deleted function %s\n", functionName)

		// Delete log group if requested
		if input.DeleteLogs {
			logGroup := loggroup.NewLogGroupForFunction(functionName)
			fmt.Fprintf(uc.output, "Deleting CloudWatch Logs log group %s...\n", logGroup.Name())

			if err := uc.logGroupRepo.Delete(ctx, logGroup); err != nil {
				fmt.Fprintf(uc.output, "Warning: Failed to delete log group: %v\n", err)
				// Don't count this as a failure since the function was deleted
			} else {
				fmt.Fprintf(uc.output, "Deleted CloudWatch Logs log group %s\n", logGroup.Name())
			}
		} else {
			fmt.Fprintf(uc.output, "Skipping log deletion (--without-logs specified)\n")
		}

		fmt.Fprintf(uc.output, "Successfully processed %s\n", functionName)
		successCount++
	}

	fmt.Fprintf(uc.output, "\n=== Summary ===\n")
	fmt.Fprintf(uc.output, "Total functions: %d\n", len(functionNames))
	fmt.Fprintf(uc.output, "Successfully deleted: %d\n", successCount)
	fmt.Fprintf(uc.output, "Failed: %d\n", failureCount)

	if failureCount > 0 {
		return fmt.Errorf("failed to delete %d function(s)", failureCount)
	}

	return nil
}

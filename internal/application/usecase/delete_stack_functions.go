package usecase

import (
	"context"
	"fmt"

	"github.com/shirasu/clambda/internal/domain/function"
	"github.com/shirasu/clambda/internal/domain/loggroup"
	"github.com/shirasu/clambda/internal/domain/stack"
)

// DeleteStackFunctionsUseCase handles deleting all Lambda functions in a stack
type DeleteStackFunctionsUseCase struct {
	functionRepo function.Repository
	logGroupRepo loggroup.Repository
	stackRepo    stack.Repository
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
) *DeleteStackFunctionsUseCase {
	return &DeleteStackFunctionsUseCase{
		functionRepo: functionRepo,
		logGroupRepo: logGroupRepo,
		stackRepo:    stackRepo,
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

	fmt.Printf("Found %d Lambda function(s) in stack %s\n", len(functionNames), input.StackName)

	// Delete each function
	successCount := 0
	failureCount := 0
	for _, functionName := range functionNames {
		fmt.Printf("\n=== Processing function: %s ===\n", functionName)

		// Get the function to check VPC status
		fn, err := uc.functionRepo.FindByName(ctx, functionName)
		if err != nil {
			fmt.Printf("❌ Failed to get function: %v\n", err)
			failureCount++
			continue
		}

		// Handle VPC detachment if requested
		if input.DetachVPC && fn.IsAttachedToVPC() {
			// Disable IPv6 if requested and enabled
			if input.DisableIPv6 && fn.HasIPv6Enabled() {
				fmt.Printf("Disabling IPv6...\n")
				if err := uc.functionRepo.DisableIPv6(ctx, functionName); err != nil {
					fmt.Printf("❌ Failed to disable IPv6: %v\n", err)
					failureCount++
					continue
				}
				fmt.Printf("✓ IPv6 disabled\n")
			}

			// Detach VPC
			fmt.Printf("Detaching VPC...\n")
			if err := uc.functionRepo.DetachVPC(ctx, functionName); err != nil {
				fmt.Printf("❌ Failed to detach VPC: %v\n", err)
				failureCount++
				continue
			}
			fmt.Printf("✓ VPC detached\n")
		}

		// Delete the function
		fmt.Printf("Deleting Lambda function...\n")
		if err := uc.functionRepo.Delete(ctx, functionName); err != nil {
			fmt.Printf("❌ Failed to delete function: %v\n", err)
			failureCount++
			continue
		}
		fmt.Printf("✓ Lambda function deleted\n")

		// Delete log group if requested
		if input.DeleteLogs {
			logGroupName := fmt.Sprintf("/aws/lambda/%s", functionName)
			fmt.Printf("Deleting CloudWatch Logs log group %s...\n", logGroupName)

			logGroup := loggroup.NewLogGroup(logGroupName)
			if err := uc.logGroupRepo.Delete(ctx, logGroup); err != nil {
				fmt.Printf("⚠️  Warning: Failed to delete log group: %v\n", err)
				// Don't count this as a failure since the function was deleted
			} else {
				fmt.Printf("✓ Log group deleted\n")
			}
		}

		fmt.Printf("✓ Successfully processed %s\n", functionName)
		successCount++
	}

	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Total functions: %d\n", len(functionNames))
	fmt.Printf("Successfully deleted: %d\n", successCount)
	fmt.Printf("Failed: %d\n", failureCount)

	if failureCount > 0 {
		return fmt.Errorf("failed to delete %d function(s)", failureCount)
	}

	return nil
}

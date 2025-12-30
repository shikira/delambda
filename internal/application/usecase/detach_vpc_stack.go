package usecase

import (
	"context"
	"fmt"

	"github.com/shirasu/clambda/internal/domain/function"
	"github.com/shirasu/clambda/internal/domain/stack"
)

// DetachVPCStackUseCase handles detaching VPC from all Lambda functions in a stack
type DetachVPCStackUseCase struct {
	functionRepo function.Repository
	stackRepo    stack.Repository
}

// DetachVPCStackInput contains the input parameters for detaching VPC from stack functions
type DetachVPCStackInput struct {
	StackName   string
	DisableIPv6 bool
}

// NewDetachVPCStackUseCase creates a new DetachVPCStackUseCase
func NewDetachVPCStackUseCase(functionRepo function.Repository, stackRepo stack.Repository) *DetachVPCStackUseCase {
	return &DetachVPCStackUseCase{
		functionRepo: functionRepo,
		stackRepo:    stackRepo,
	}
}

// Execute detaches VPC from all Lambda functions in the specified stack
func (uc *DetachVPCStackUseCase) Execute(ctx context.Context, input *DetachVPCStackInput) error {
	// Get all Lambda functions in the stack
	functionNames, err := uc.stackRepo.ListLambdaFunctions(ctx, input.StackName)
	if err != nil {
		return fmt.Errorf("failed to list Lambda functions in stack: %w", err)
	}

	if len(functionNames) == 0 {
		return fmt.Errorf("no Lambda functions found in stack %s", input.StackName)
	}

	fmt.Printf("Found %d Lambda function(s) in stack %s\n", len(functionNames), input.StackName)

	// Detach VPC from each function
	successCount := 0
	failureCount := 0
	for _, functionName := range functionNames {
		fmt.Printf("\nProcessing function: %s\n", functionName)

		// Get the function
		fn, err := uc.functionRepo.FindByName(ctx, functionName)
		if err != nil {
			fmt.Printf("  ❌ Failed to get function: %v\n", err)
			failureCount++
			continue
		}

		// Check if function has VPC
		if !fn.IsAttachedToVPC() {
			fmt.Printf("  ⏭️  Function is not attached to VPC, skipping\n")
			successCount++
			continue
		}

		// Disable IPv6 if requested and enabled
		if input.DisableIPv6 && fn.HasIPv6Enabled() {
			fmt.Printf("  Disabling IPv6...\n")
			if err := uc.functionRepo.DisableIPv6(ctx, functionName); err != nil {
				fmt.Printf("  ❌ Failed to disable IPv6: %v\n", err)
				failureCount++
				continue
			}
			fmt.Printf("  ✓ IPv6 disabled\n")
		}

		// Detach VPC
		fmt.Printf("  Detaching VPC...\n")
		if err := uc.functionRepo.DetachVPC(ctx, functionName); err != nil {
			fmt.Printf("  ❌ Failed to detach VPC: %v\n", err)
			failureCount++
			continue
		}

		fmt.Printf("  ✓ VPC detached successfully\n")
		successCount++
	}

	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Total functions: %d\n", len(functionNames))
	fmt.Printf("Successfully processed: %d\n", successCount)
	fmt.Printf("Failed: %d\n", failureCount)

	if failureCount > 0 {
		return fmt.Errorf("failed to process %d function(s)", failureCount)
	}

	return nil
}

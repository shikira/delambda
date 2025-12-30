package usecase

import (
	"context"
	"fmt"

	"github.com/shirasu/clambda/internal/domain/function"
)

// DetachVPCUseCase handles detaching VPC from Lambda functions
type DetachVPCUseCase struct {
	functionRepo function.Repository
}

// DetachVPCInput represents the input for detaching VPC
type DetachVPCInput struct {
	FunctionName string
	DisableIPv6  bool
}

// NewDetachVPCUseCase creates a new DetachVPCUseCase
func NewDetachVPCUseCase(functionRepo function.Repository) *DetachVPCUseCase {
	return &DetachVPCUseCase{
		functionRepo: functionRepo,
	}
}

// Execute executes the detach VPC use case
func (uc *DetachVPCUseCase) Execute(ctx context.Context, input *DetachVPCInput) error {
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
		return fmt.Errorf("failed to detach VPC: %w", err)
	}

	return nil
}

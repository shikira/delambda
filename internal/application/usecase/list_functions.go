package usecase

import (
	"context"

	"github.com/shirasu/clambda/internal/domain/function"
)

// ListFunctionsUseCase handles listing Lambda functions
type ListFunctionsUseCase struct {
	functionRepo function.Repository
}

// NewListFunctionsUseCase creates a new ListFunctionsUseCase
func NewListFunctionsUseCase(functionRepo function.Repository) *ListFunctionsUseCase {
	return &ListFunctionsUseCase{
		functionRepo: functionRepo,
	}
}

// Execute executes the list functions use case
func (uc *ListFunctionsUseCase) Execute(ctx context.Context) ([]*function.Function, error) {
	return uc.functionRepo.FindAll(ctx)
}

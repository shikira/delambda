package usecase

import (
	"context"

	"github.com/shirasu/clambda/internal/domain/loggroup"
)

// DeleteLogGroupUseCase handles deleting log groups
type DeleteLogGroupUseCase struct {
	logGroupRepo loggroup.Repository
}

// NewDeleteLogGroupUseCase creates a new DeleteLogGroupUseCase
func NewDeleteLogGroupUseCase(logGroupRepo loggroup.Repository) *DeleteLogGroupUseCase {
	return &DeleteLogGroupUseCase{
		logGroupRepo: logGroupRepo,
	}
}

// Execute executes the delete log group use case
func (uc *DeleteLogGroupUseCase) Execute(ctx context.Context, functionName string) error {
	logGroup := loggroup.NewLogGroupForFunction(functionName)
	return uc.logGroupRepo.Delete(ctx, logGroup)
}

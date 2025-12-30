package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/shirasu/clambda/internal/domain/loggroup"
	logspkg "github.com/shirasu/clambda/internal/logs"
)

// LogGroupRepository implements the loggroup.Repository interface
type LogGroupRepository struct {
	client logspkg.LogsAPI
}

// NewLogGroupRepository creates a new LogGroupRepository
func NewLogGroupRepository(client logspkg.LogsAPI) *LogGroupRepository {
	return &LogGroupRepository{
		client: client,
	}
}

// Exists checks if a log group exists
func (r *LogGroupRepository) Exists(ctx context.Context, logGroup *loggroup.LogGroup) (bool, error) {
	output, err := r.client.DescribeLogGroups(ctx, &cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupNamePrefix: aws.String(logGroup.Name()),
	})
	if err != nil {
		return false, fmt.Errorf("failed to describe log groups: %w", err)
	}

	for _, lg := range output.LogGroups {
		if aws.ToString(lg.LogGroupName) == logGroup.Name() {
			return true, nil
		}
	}

	return false, nil
}

// Delete deletes a log group
func (r *LogGroupRepository) Delete(ctx context.Context, logGroup *loggroup.LogGroup) error {
	_, err := r.client.DeleteLogGroup(ctx, &cloudwatchlogs.DeleteLogGroupInput{
		LogGroupName: aws.String(logGroup.Name()),
	})
	if err != nil {
		// Check if the error is because the log group doesn't exist
		if strings.Contains(err.Error(), "ResourceNotFoundException") {
			return nil // Log group already deleted
		}
		return fmt.Errorf("failed to delete log group %s: %w", logGroup.Name(), err)
	}
	return nil
}

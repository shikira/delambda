package logs

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
)

// LogsAPI defines the interface for CloudWatch Logs operations
type LogsAPI interface {
	DescribeLogGroups(ctx context.Context, params *cloudwatchlogs.DescribeLogGroupsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogGroupsOutput, error)
	DeleteLogGroup(ctx context.Context, params *cloudwatchlogs.DeleteLogGroupInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DeleteLogGroupOutput, error)
}

// Service handles CloudWatch Logs operations
type Service struct {
	client LogsAPI
}

// NewService creates a new Logs service
func NewService(client *cloudwatchlogs.Client) *Service {
	return &Service{
		client: client,
	}
}

// GetLogGroupName returns the CloudWatch Logs group name for a Lambda function
func GetLogGroupName(functionName string) string {
	return fmt.Sprintf("/aws/lambda/%s", functionName)
}

// LogGroupExists checks if a log group exists
func (s *Service) LogGroupExists(ctx context.Context, logGroupName string) (bool, error) {
	output, err := s.client.DescribeLogGroups(ctx, &cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupNamePrefix: aws.String(logGroupName),
	})
	if err != nil {
		return false, fmt.Errorf("failed to describe log groups: %w", err)
	}

	for _, lg := range output.LogGroups {
		if aws.ToString(lg.LogGroupName) == logGroupName {
			return true, nil
		}
	}

	return false, nil
}

// DeleteLogGroup deletes a CloudWatch Logs group
func (s *Service) DeleteLogGroup(ctx context.Context, logGroupName string) error {
	_, err := s.client.DeleteLogGroup(ctx, &cloudwatchlogs.DeleteLogGroupInput{
		LogGroupName: aws.String(logGroupName),
	})
	if err != nil {
		// Check if the error is because the log group doesn't exist
		if strings.Contains(err.Error(), "ResourceNotFoundException") {
			return nil // Log group already deleted
		}
		return fmt.Errorf("failed to delete log group %s: %w", logGroupName, err)
	}
	return nil
}

// DeleteFunctionLogGroup deletes the CloudWatch Logs group for a Lambda function
func (s *Service) DeleteFunctionLogGroup(ctx context.Context, functionName string) error {
	logGroupName := GetLogGroupName(functionName)
	return s.DeleteLogGroup(ctx, logGroupName)
}

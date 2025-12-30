package repository

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
)

// StackRepository implements the stack repository using AWS SDK
type StackRepository struct {
	client *cloudformation.Client
}

// NewStackRepository creates a new stack repository
func NewStackRepository(client *cloudformation.Client) *StackRepository {
	return &StackRepository{
		client: client,
	}
}

// ListLambdaFunctions returns all Lambda function names in the specified stack
func (r *StackRepository) ListLambdaFunctions(ctx context.Context, stackName string) ([]string, error) {
	// Get stack resources
	input := &cloudformation.ListStackResourcesInput{
		StackName: &stackName,
	}

	var functionNames []string
	paginator := cloudformation.NewListStackResourcesPaginator(r.client, input)

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list stack resources: %w", err)
		}

		// Filter for Lambda functions
		for _, resource := range output.StackResourceSummaries {
			if resource.ResourceType != nil && *resource.ResourceType == "AWS::Lambda::Function" {
				if resource.PhysicalResourceId != nil {
					functionNames = append(functionNames, *resource.PhysicalResourceId)
				}
			}
		}
	}

	return functionNames, nil
}

// StackExists checks if a stack exists
func (r *StackRepository) StackExists(ctx context.Context, stackName string) (bool, error) {
	input := &cloudformation.DescribeStacksInput{
		StackName: &stackName,
	}

	_, err := r.client.DescribeStacks(ctx, input)
	if err != nil {
		// Check if error is because stack doesn't exist
		if err != nil && err.Error() != "" {
			return false, nil
		}
		return false, fmt.Errorf("failed to describe stack: %w", err)
	}

	return true, nil
}

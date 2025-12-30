package lambda

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

// LambdaAPI defines the interface for Lambda operations
type LambdaAPI interface {
	GetFunction(ctx context.Context, params *lambda.GetFunctionInput, optFns ...func(*lambda.Options)) (*lambda.GetFunctionOutput, error)
	UpdateFunctionConfiguration(ctx context.Context, params *lambda.UpdateFunctionConfigurationInput, optFns ...func(*lambda.Options)) (*lambda.UpdateFunctionConfigurationOutput, error)
	DeleteFunction(ctx context.Context, params *lambda.DeleteFunctionInput, optFns ...func(*lambda.Options)) (*lambda.DeleteFunctionOutput, error)
	ListFunctions(ctx context.Context, params *lambda.ListFunctionsInput, optFns ...func(*lambda.Options)) (*lambda.ListFunctionsOutput, error)
}

// Service handles Lambda operations
type Service struct {
	client LambdaAPI
}

// NewService creates a new Lambda service
func NewService(client *lambda.Client) *Service {
	return &Service{
		client: client,
	}
}

// ListFunctions lists all Lambda functions
func (s *Service) ListFunctions(ctx context.Context) ([]types.FunctionConfiguration, error) {
	var functions []types.FunctionConfiguration
	var nextMarker *string

	for {
		input := &lambda.ListFunctionsInput{
			Marker: nextMarker,
		}

		output, err := s.client.ListFunctions(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to list functions: %w", err)
		}

		functions = append(functions, output.Functions...)

		if output.NextMarker == nil {
			break
		}
		nextMarker = output.NextMarker
	}

	return functions, nil
}

// GetFunction retrieves function configuration
func (s *Service) GetFunction(ctx context.Context, functionName string) (*lambda.GetFunctionOutput, error) {
	output, err := s.client.GetFunction(ctx, &lambda.GetFunctionInput{
		FunctionName: aws.String(functionName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get function %s: %w", functionName, err)
	}
	return output, nil
}

// DisableIPv6 disables IPv6 for the Lambda function's VPC configuration
func (s *Service) DisableIPv6(ctx context.Context, functionName string) error {
	// Get current configuration
	fn, err := s.GetFunction(ctx, functionName)
	if err != nil {
		return err
	}

	if fn.Configuration.VpcConfig == nil {
		return fmt.Errorf("function %s is not attached to a VPC", functionName)
	}

	// Update VPC configuration to disable IPv6
	_, err = s.client.UpdateFunctionConfiguration(ctx, &lambda.UpdateFunctionConfigurationInput{
		FunctionName: aws.String(functionName),
		VpcConfig: &types.VpcConfig{
			SubnetIds:               fn.Configuration.VpcConfig.SubnetIds,
			SecurityGroupIds:        fn.Configuration.VpcConfig.SecurityGroupIds,
			Ipv6AllowedForDualStack: aws.Bool(false),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to disable IPv6 for function %s: %w", functionName, err)
	}

	// Wait for the update to complete
	return s.waitForFunctionUpdate(ctx, functionName)
}

// DetachVPC detaches the VPC from the Lambda function
func (s *Service) DetachVPC(ctx context.Context, functionName string) error {
	// Get current configuration to verify VPC is attached
	fn, err := s.GetFunction(ctx, functionName)
	if err != nil {
		return err
	}

	if fn.Configuration.VpcConfig == nil || len(fn.Configuration.VpcConfig.SubnetIds) == 0 {
		return fmt.Errorf("function %s is not attached to a VPC", functionName)
	}

	// Update function configuration to remove VPC
	_, err = s.client.UpdateFunctionConfiguration(ctx, &lambda.UpdateFunctionConfigurationInput{
		FunctionName: aws.String(functionName),
		VpcConfig: &types.VpcConfig{
			SubnetIds:        []string{},
			SecurityGroupIds: []string{},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to detach VPC from function %s: %w", functionName, err)
	}

	// Wait for the update to complete
	return s.waitForFunctionUpdate(ctx, functionName)
}

// DeleteFunction deletes a Lambda function
func (s *Service) DeleteFunction(ctx context.Context, functionName string) error {
	_, err := s.client.DeleteFunction(ctx, &lambda.DeleteFunctionInput{
		FunctionName: aws.String(functionName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete function %s: %w", functionName, err)
	}
	return nil
}

// waitForFunctionUpdate waits for the function to be in Active state
func (s *Service) waitForFunctionUpdate(ctx context.Context, functionName string) error {
	maxAttempts := 60
	interval := 5 * time.Second

	for i := 0; i < maxAttempts; i++ {
		fn, err := s.GetFunction(ctx, functionName)
		if err != nil {
			return err
		}

		state := fn.Configuration.State
		lastUpdateStatus := fn.Configuration.LastUpdateStatus

		// Check if function is ready
		if state == types.StateActive && lastUpdateStatus == types.LastUpdateStatusSuccessful {
			return nil
		}

		// Check for failed state
		if state == types.StateFailed || lastUpdateStatus == types.LastUpdateStatusFailed {
			reason := string(fn.Configuration.StateReasonCode)
			return fmt.Errorf("function update failed: state=%s, lastUpdateStatus=%s, reason=%s",
				state, lastUpdateStatus, reason)
		}

		time.Sleep(interval)
	}

	return fmt.Errorf("timeout waiting for function %s to be ready", functionName)
}

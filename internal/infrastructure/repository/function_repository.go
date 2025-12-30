package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/shirasu/clambda/internal/domain/function"
	lambdapkg "github.com/shirasu/clambda/internal/lambda"
)

// FunctionRepository implements the function.Repository interface
type FunctionRepository struct {
	client lambdapkg.LambdaAPI
}

// NewFunctionRepository creates a new FunctionRepository
func NewFunctionRepository(client lambdapkg.LambdaAPI) *FunctionRepository {
	return &FunctionRepository{
		client: client,
	}
}

// FindAll returns all Lambda functions
func (r *FunctionRepository) FindAll(ctx context.Context) ([]*function.Function, error) {
	var functions []*function.Function
	var nextMarker *string

	for {
		input := &lambda.ListFunctionsInput{
			Marker: nextMarker,
		}

		output, err := r.client.ListFunctions(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to list functions: %w", err)
		}

		for _, fn := range output.Functions {
			var vpcConfig *function.VPCConfig
			if fn.VpcConfig != nil {
				vpcConfig = &function.VPCConfig{
					VPCId:                   aws.ToString(fn.VpcConfig.VpcId),
					SubnetIds:               fn.VpcConfig.SubnetIds,
					SecurityGroupIds:        fn.VpcConfig.SecurityGroupIds,
					IPv6AllowedForDualStack: aws.ToBool(fn.VpcConfig.Ipv6AllowedForDualStack),
				}
			}
			functions = append(functions, function.NewFunction(
				aws.ToString(fn.FunctionName),
				fn.Runtime,
				fn.State,
				vpcConfig,
			))
		}

		if output.NextMarker == nil {
			break
		}
		nextMarker = output.NextMarker
	}

	return functions, nil
}

// FindByName finds a Lambda function by name
func (r *FunctionRepository) FindByName(ctx context.Context, name string) (*function.Function, error) {
	output, err := r.client.GetFunction(ctx, &lambda.GetFunctionInput{
		FunctionName: aws.String(name),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get function %s: %w", name, err)
	}

	var vpcConfig *function.VPCConfig
	if output.Configuration.VpcConfig != nil {
		vpcConfig = &function.VPCConfig{
			VPCId:                   aws.ToString(output.Configuration.VpcConfig.VpcId),
			SubnetIds:               output.Configuration.VpcConfig.SubnetIds,
			SecurityGroupIds:        output.Configuration.VpcConfig.SecurityGroupIds,
			IPv6AllowedForDualStack: aws.ToBool(output.Configuration.VpcConfig.Ipv6AllowedForDualStack),
		}
	}

	return function.NewFunction(
		aws.ToString(output.Configuration.FunctionName),
		output.Configuration.Runtime,
		output.Configuration.State,
		vpcConfig,
	), nil
}

// DisableIPv6 disables IPv6 for a Lambda function
func (r *FunctionRepository) DisableIPv6(ctx context.Context, functionName string) error {
	// Get current configuration
	fn, err := r.FindByName(ctx, functionName)
	if err != nil {
		return err
	}

	if !fn.IsAttachedToVPC() {
		return fmt.Errorf("function %s is not attached to a VPC", functionName)
	}

	vpcConfig := fn.VPCConfig()

	// Update VPC configuration to disable IPv6
	_, err = r.client.UpdateFunctionConfiguration(ctx, &lambda.UpdateFunctionConfigurationInput{
		FunctionName: aws.String(functionName),
		VpcConfig: &types.VpcConfig{
			SubnetIds:               vpcConfig.SubnetIds,
			SecurityGroupIds:        vpcConfig.SecurityGroupIds,
			Ipv6AllowedForDualStack: aws.Bool(false),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to disable IPv6 for function %s: %w", functionName, err)
	}

	// Wait for the update to complete
	return r.waitForFunctionUpdate(ctx, functionName)
}

// DetachVPC detaches VPC from a Lambda function
func (r *FunctionRepository) DetachVPC(ctx context.Context, functionName string) error {
	// Get current configuration to verify VPC is attached
	fn, err := r.FindByName(ctx, functionName)
	if err != nil {
		return err
	}

	if !fn.IsAttachedToVPC() {
		return fmt.Errorf("function %s is not attached to a VPC", functionName)
	}

	// Update function configuration to remove VPC
	_, err = r.client.UpdateFunctionConfiguration(ctx, &lambda.UpdateFunctionConfigurationInput{
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
	return r.waitForFunctionUpdate(ctx, functionName)
}

// Delete deletes a Lambda function
func (r *FunctionRepository) Delete(ctx context.Context, functionName string) error {
	_, err := r.client.DeleteFunction(ctx, &lambda.DeleteFunctionInput{
		FunctionName: aws.String(functionName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete function %s: %w", functionName, err)
	}
	return nil
}

// waitForFunctionUpdate waits for the function to be in Active state
func (r *FunctionRepository) waitForFunctionUpdate(ctx context.Context, functionName string) error {
	maxAttempts := 60
	interval := 5 * time.Second

	for i := 0; i < maxAttempts; i++ {
		output, err := r.client.GetFunction(ctx, &lambda.GetFunctionInput{
			FunctionName: aws.String(functionName),
		})
		if err != nil {
			return err
		}

		state := output.Configuration.State
		lastUpdateStatus := output.Configuration.LastUpdateStatus

		// Check if function is ready
		if state == types.StateActive && lastUpdateStatus == types.LastUpdateStatusSuccessful {
			return nil
		}

		// Check for failed state
		if state == types.StateFailed || lastUpdateStatus == types.LastUpdateStatusFailed {
			reason := string(output.Configuration.StateReasonCode)
			return fmt.Errorf("function update failed: state=%s, lastUpdateStatus=%s, reason=%s",
				state, lastUpdateStatus, reason)
		}

		time.Sleep(interval)
	}

	return fmt.Errorf("timeout waiting for function %s to be ready", functionName)
}

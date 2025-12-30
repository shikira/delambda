package lambda

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

// Mock Lambda Client
type mockLambdaClient struct {
	listFunctionsFunc               func(ctx context.Context, params *lambda.ListFunctionsInput, optFns ...func(*lambda.Options)) (*lambda.ListFunctionsOutput, error)
	getFunctionFunc                 func(ctx context.Context, params *lambda.GetFunctionInput, optFns ...func(*lambda.Options)) (*lambda.GetFunctionOutput, error)
	updateFunctionConfigurationFunc func(ctx context.Context, params *lambda.UpdateFunctionConfigurationInput, optFns ...func(*lambda.Options)) (*lambda.UpdateFunctionConfigurationOutput, error)
	deleteFunctionFunc              func(ctx context.Context, params *lambda.DeleteFunctionInput, optFns ...func(*lambda.Options)) (*lambda.DeleteFunctionOutput, error)
}

func (m *mockLambdaClient) ListFunctions(ctx context.Context, params *lambda.ListFunctionsInput, optFns ...func(*lambda.Options)) (*lambda.ListFunctionsOutput, error) {
	if m.listFunctionsFunc != nil {
		return m.listFunctionsFunc(ctx, params, optFns...)
	}
	return &lambda.ListFunctionsOutput{}, nil
}

func (m *mockLambdaClient) GetFunction(ctx context.Context, params *lambda.GetFunctionInput, optFns ...func(*lambda.Options)) (*lambda.GetFunctionOutput, error) {
	if m.getFunctionFunc != nil {
		return m.getFunctionFunc(ctx, params, optFns...)
	}
	return &lambda.GetFunctionOutput{}, nil
}

func (m *mockLambdaClient) UpdateFunctionConfiguration(ctx context.Context, params *lambda.UpdateFunctionConfigurationInput, optFns ...func(*lambda.Options)) (*lambda.UpdateFunctionConfigurationOutput, error) {
	if m.updateFunctionConfigurationFunc != nil {
		return m.updateFunctionConfigurationFunc(ctx, params, optFns...)
	}
	return &lambda.UpdateFunctionConfigurationOutput{}, nil
}

func (m *mockLambdaClient) DeleteFunction(ctx context.Context, params *lambda.DeleteFunctionInput, optFns ...func(*lambda.Options)) (*lambda.DeleteFunctionOutput, error) {
	if m.deleteFunctionFunc != nil {
		return m.deleteFunctionFunc(ctx, params, optFns...)
	}
	return &lambda.DeleteFunctionOutput{}, nil
}

func TestListFunctions(t *testing.T) {
	tests := []struct {
		name      string
		mockFunc  func(ctx context.Context, params *lambda.ListFunctionsInput, optFns ...func(*lambda.Options)) (*lambda.ListFunctionsOutput, error)
		wantCount int
		wantErr   bool
	}{
		{
			name: "successful list with multiple functions",
			mockFunc: func(ctx context.Context, params *lambda.ListFunctionsInput, optFns ...func(*lambda.Options)) (*lambda.ListFunctionsOutput, error) {
				return &lambda.ListFunctionsOutput{
					Functions: []types.FunctionConfiguration{
						{FunctionName: aws.String("func1")},
						{FunctionName: aws.String("func2")},
					},
				}, nil
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "empty list",
			mockFunc: func(ctx context.Context, params *lambda.ListFunctionsInput, optFns ...func(*lambda.Options)) (*lambda.ListFunctionsOutput, error) {
				return &lambda.ListFunctionsOutput{
					Functions: []types.FunctionConfiguration{},
				}, nil
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "error listing functions",
			mockFunc: func(ctx context.Context, params *lambda.ListFunctionsInput, optFns ...func(*lambda.Options)) (*lambda.ListFunctionsOutput, error) {
				return nil, errors.New("API error")
			},
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockLambdaClient{
				listFunctionsFunc: tt.mockFunc,
			}
			svc := &Service{client: mock}

			functions, err := svc.ListFunctions(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("ListFunctions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(functions) != tt.wantCount {
				t.Errorf("ListFunctions() got %d functions, want %d", len(functions), tt.wantCount)
			}
		})
	}
}

func TestGetFunction(t *testing.T) {
	tests := []struct {
		name         string
		functionName string
		mockFunc     func(ctx context.Context, params *lambda.GetFunctionInput, optFns ...func(*lambda.Options)) (*lambda.GetFunctionOutput, error)
		wantErr      bool
	}{
		{
			name:         "successful get",
			functionName: "test-function",
			mockFunc: func(ctx context.Context, params *lambda.GetFunctionInput, optFns ...func(*lambda.Options)) (*lambda.GetFunctionOutput, error) {
				return &lambda.GetFunctionOutput{
					Configuration: &types.FunctionConfiguration{
						FunctionName: aws.String("test-function"),
					},
				}, nil
			},
			wantErr: false,
		},
		{
			name:         "function not found",
			functionName: "non-existent",
			mockFunc: func(ctx context.Context, params *lambda.GetFunctionInput, optFns ...func(*lambda.Options)) (*lambda.GetFunctionOutput, error) {
				return nil, errors.New("ResourceNotFoundException")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockLambdaClient{
				getFunctionFunc: tt.mockFunc,
			}
			svc := &Service{client: mock}

			_, err := svc.GetFunction(context.Background(), tt.functionName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFunction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDisableIPv6(t *testing.T) {
	tests := []struct {
		name               string
		functionName       string
		getFunctionMock    func(ctx context.Context, params *lambda.GetFunctionInput, optFns ...func(*lambda.Options)) (*lambda.GetFunctionOutput, error)
		updateFunctionMock func(ctx context.Context, params *lambda.UpdateFunctionConfigurationInput, optFns ...func(*lambda.Options)) (*lambda.UpdateFunctionConfigurationOutput, error)
		wantErr            bool
	}{
		{
			name:         "successful IPv6 disable",
			functionName: "test-function",
			getFunctionMock: (func() func(ctx context.Context, params *lambda.GetFunctionInput, optFns ...func(*lambda.Options)) (*lambda.GetFunctionOutput, error) {
				callCount := 0
				return func(ctx context.Context, params *lambda.GetFunctionInput, optFns ...func(*lambda.Options)) (*lambda.GetFunctionOutput, error) {
					callCount++
					state := types.StateActive
					status := types.LastUpdateStatusSuccessful
					if callCount == 1 {
						status = types.LastUpdateStatusInProgress
					}
					return &lambda.GetFunctionOutput{
						Configuration: &types.FunctionConfiguration{
							FunctionName: aws.String("test-function"),
							VpcConfig: &types.VpcConfigResponse{
								SubnetIds:               []string{"subnet-1"},
								SecurityGroupIds:        []string{"sg-1"},
								Ipv6AllowedForDualStack: aws.Bool(true),
							},
							State:            state,
							LastUpdateStatus: status,
						},
					}, nil
				}
			})(),
			updateFunctionMock: func(ctx context.Context, params *lambda.UpdateFunctionConfigurationInput, optFns ...func(*lambda.Options)) (*lambda.UpdateFunctionConfigurationOutput, error) {
				return &lambda.UpdateFunctionConfigurationOutput{}, nil
			},
			wantErr: false,
		},
		{
			name:         "function not attached to VPC",
			functionName: "test-function",
			getFunctionMock: func(ctx context.Context, params *lambda.GetFunctionInput, optFns ...func(*lambda.Options)) (*lambda.GetFunctionOutput, error) {
				return &lambda.GetFunctionOutput{
					Configuration: &types.FunctionConfiguration{
						FunctionName: aws.String("test-function"),
						VpcConfig:    nil,
					},
				}, nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockLambdaClient{
				getFunctionFunc:                 tt.getFunctionMock,
				updateFunctionConfigurationFunc: tt.updateFunctionMock,
			}
			svc := &Service{client: mock}

			err := svc.DisableIPv6(context.Background(), tt.functionName)
			if (err != nil) != tt.wantErr {
				t.Errorf("DisableIPv6() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDetachVPC(t *testing.T) {
	tests := []struct {
		name               string
		functionName       string
		getFunctionMock    func(ctx context.Context, params *lambda.GetFunctionInput, optFns ...func(*lambda.Options)) (*lambda.GetFunctionOutput, error)
		updateFunctionMock func(ctx context.Context, params *lambda.UpdateFunctionConfigurationInput, optFns ...func(*lambda.Options)) (*lambda.UpdateFunctionConfigurationOutput, error)
		wantErr            bool
	}{
		{
			name:         "successful VPC detach",
			functionName: "test-function",
			getFunctionMock: (func() func(ctx context.Context, params *lambda.GetFunctionInput, optFns ...func(*lambda.Options)) (*lambda.GetFunctionOutput, error) {
				callCount := 0
				return func(ctx context.Context, params *lambda.GetFunctionInput, optFns ...func(*lambda.Options)) (*lambda.GetFunctionOutput, error) {
					callCount++
					state := types.StateActive
					status := types.LastUpdateStatusSuccessful
					if callCount == 1 {
						status = types.LastUpdateStatusInProgress
					}
					return &lambda.GetFunctionOutput{
						Configuration: &types.FunctionConfiguration{
							FunctionName: aws.String("test-function"),
							VpcConfig: &types.VpcConfigResponse{
								SubnetIds:        []string{"subnet-1"},
								SecurityGroupIds: []string{"sg-1"},
							},
							State:            state,
							LastUpdateStatus: status,
						},
					}, nil
				}
			})(),
			updateFunctionMock: func(ctx context.Context, params *lambda.UpdateFunctionConfigurationInput, optFns ...func(*lambda.Options)) (*lambda.UpdateFunctionConfigurationOutput, error) {
				return &lambda.UpdateFunctionConfigurationOutput{}, nil
			},
			wantErr: false,
		},
		{
			name:         "function not attached to VPC",
			functionName: "test-function",
			getFunctionMock: func(ctx context.Context, params *lambda.GetFunctionInput, optFns ...func(*lambda.Options)) (*lambda.GetFunctionOutput, error) {
				return &lambda.GetFunctionOutput{
					Configuration: &types.FunctionConfiguration{
						FunctionName: aws.String("test-function"),
						VpcConfig: &types.VpcConfigResponse{
							SubnetIds: []string{},
						},
					},
				}, nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockLambdaClient{
				getFunctionFunc:                 tt.getFunctionMock,
				updateFunctionConfigurationFunc: tt.updateFunctionMock,
			}
			svc := &Service{client: mock}

			err := svc.DetachVPC(context.Background(), tt.functionName)
			if (err != nil) != tt.wantErr {
				t.Errorf("DetachVPC() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteFunction(t *testing.T) {
	tests := []struct {
		name         string
		functionName string
		mockFunc     func(ctx context.Context, params *lambda.DeleteFunctionInput, optFns ...func(*lambda.Options)) (*lambda.DeleteFunctionOutput, error)
		wantErr      bool
	}{
		{
			name:         "successful delete",
			functionName: "test-function",
			mockFunc: func(ctx context.Context, params *lambda.DeleteFunctionInput, optFns ...func(*lambda.Options)) (*lambda.DeleteFunctionOutput, error) {
				return &lambda.DeleteFunctionOutput{}, nil
			},
			wantErr: false,
		},
		{
			name:         "delete error",
			functionName: "test-function",
			mockFunc: func(ctx context.Context, params *lambda.DeleteFunctionInput, optFns ...func(*lambda.Options)) (*lambda.DeleteFunctionOutput, error) {
				return nil, errors.New("delete failed")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockLambdaClient{
				deleteFunctionFunc: tt.mockFunc,
			}
			svc := &Service{client: mock}

			err := svc.DeleteFunction(context.Background(), tt.functionName)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteFunction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWaitForFunctionUpdate(t *testing.T) {
	tests := []struct {
		name         string
		functionName string
		mockFunc     func(ctx context.Context, params *lambda.GetFunctionInput, optFns ...func(*lambda.Options)) (*lambda.GetFunctionOutput, error)
		wantErr      bool
	}{
		{
			name:         "immediate success",
			functionName: "test-function",
			mockFunc: func(ctx context.Context, params *lambda.GetFunctionInput, optFns ...func(*lambda.Options)) (*lambda.GetFunctionOutput, error) {
				return &lambda.GetFunctionOutput{
					Configuration: &types.FunctionConfiguration{
						State:            types.StateActive,
						LastUpdateStatus: types.LastUpdateStatusSuccessful,
					},
				}, nil
			},
			wantErr: false,
		},
		{
			name:         "failed state",
			functionName: "test-function",
			mockFunc: func(ctx context.Context, params *lambda.GetFunctionInput, optFns ...func(*lambda.Options)) (*lambda.GetFunctionOutput, error) {
				return &lambda.GetFunctionOutput{
					Configuration: &types.FunctionConfiguration{
						State:            types.StateFailed,
						LastUpdateStatus: types.LastUpdateStatusFailed,
						StateReasonCode:  types.StateReasonCodeInternalError,
					},
				}, nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockLambdaClient{
				getFunctionFunc: tt.mockFunc,
			}
			svc := &Service{client: mock}

			// Override the sleep interval for testing
			originalInterval := 5 * time.Second
			defer func() { _ = originalInterval }()

			err := svc.waitForFunctionUpdate(context.Background(), tt.functionName)
			if (err != nil) != tt.wantErr {
				t.Errorf("waitForFunctionUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

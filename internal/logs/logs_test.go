package logs

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

// Mock CloudWatch Logs Client
type mockLogsClient struct {
	describeLogGroupsFunc func(ctx context.Context, params *cloudwatchlogs.DescribeLogGroupsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogGroupsOutput, error)
	deleteLogGroupFunc    func(ctx context.Context, params *cloudwatchlogs.DeleteLogGroupInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DeleteLogGroupOutput, error)
}

func (m *mockLogsClient) DescribeLogGroups(ctx context.Context, params *cloudwatchlogs.DescribeLogGroupsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogGroupsOutput, error) {
	if m.describeLogGroupsFunc != nil {
		return m.describeLogGroupsFunc(ctx, params, optFns...)
	}
	return &cloudwatchlogs.DescribeLogGroupsOutput{}, nil
}

func (m *mockLogsClient) DeleteLogGroup(ctx context.Context, params *cloudwatchlogs.DeleteLogGroupInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DeleteLogGroupOutput, error) {
	if m.deleteLogGroupFunc != nil {
		return m.deleteLogGroupFunc(ctx, params, optFns...)
	}
	return &cloudwatchlogs.DeleteLogGroupOutput{}, nil
}

func TestGetLogGroupName(t *testing.T) {
	tests := []struct {
		name         string
		functionName string
		want         string
	}{
		{
			name:         "simple function name",
			functionName: "my-function",
			want:         "/aws/lambda/my-function",
		},
		{
			name:         "function name with hyphens",
			functionName: "my-test-function",
			want:         "/aws/lambda/my-test-function",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetLogGroupName(tt.functionName); got != tt.want {
				t.Errorf("GetLogGroupName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLogGroupExists(t *testing.T) {
	tests := []struct {
		name         string
		logGroupName string
		mockFunc     func(ctx context.Context, params *cloudwatchlogs.DescribeLogGroupsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogGroupsOutput, error)
		want         bool
		wantErr      bool
	}{
		{
			name:         "log group exists",
			logGroupName: "/aws/lambda/test-function",
			mockFunc: func(ctx context.Context, params *cloudwatchlogs.DescribeLogGroupsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogGroupsOutput, error) {
				return &cloudwatchlogs.DescribeLogGroupsOutput{
					LogGroups: []types.LogGroup{
						{
							LogGroupName: aws.String("/aws/lambda/test-function"),
						},
					},
				}, nil
			},
			want:    true,
			wantErr: false,
		},
		{
			name:         "log group does not exist",
			logGroupName: "/aws/lambda/non-existent",
			mockFunc: func(ctx context.Context, params *cloudwatchlogs.DescribeLogGroupsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogGroupsOutput, error) {
				return &cloudwatchlogs.DescribeLogGroupsOutput{
					LogGroups: []types.LogGroup{},
				}, nil
			},
			want:    false,
			wantErr: false,
		},
		{
			name:         "API error",
			logGroupName: "/aws/lambda/test-function",
			mockFunc: func(ctx context.Context, params *cloudwatchlogs.DescribeLogGroupsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogGroupsOutput, error) {
				return nil, errors.New("API error")
			},
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockLogsClient{
				describeLogGroupsFunc: tt.mockFunc,
			}
			svc := &Service{client: mock}

			got, err := svc.LogGroupExists(context.Background(), tt.logGroupName)
			if (err != nil) != tt.wantErr {
				t.Errorf("LogGroupExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("LogGroupExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeleteLogGroup(t *testing.T) {
	tests := []struct {
		name         string
		logGroupName string
		mockFunc     func(ctx context.Context, params *cloudwatchlogs.DeleteLogGroupInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DeleteLogGroupOutput, error)
		wantErr      bool
	}{
		{
			name:         "successful delete",
			logGroupName: "/aws/lambda/test-function",
			mockFunc: func(ctx context.Context, params *cloudwatchlogs.DeleteLogGroupInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DeleteLogGroupOutput, error) {
				return &cloudwatchlogs.DeleteLogGroupOutput{}, nil
			},
			wantErr: false,
		},
		{
			name:         "log group not found (should not error)",
			logGroupName: "/aws/lambda/non-existent",
			mockFunc: func(ctx context.Context, params *cloudwatchlogs.DeleteLogGroupInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DeleteLogGroupOutput, error) {
				return nil, errors.New("ResourceNotFoundException: log group not found")
			},
			wantErr: false,
		},
		{
			name:         "other error",
			logGroupName: "/aws/lambda/test-function",
			mockFunc: func(ctx context.Context, params *cloudwatchlogs.DeleteLogGroupInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DeleteLogGroupOutput, error) {
				return nil, errors.New("InternalError")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockLogsClient{
				deleteLogGroupFunc: tt.mockFunc,
			}
			svc := &Service{client: mock}

			err := svc.DeleteLogGroup(context.Background(), tt.logGroupName)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteLogGroup() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteFunctionLogGroup(t *testing.T) {
	tests := []struct {
		name         string
		functionName string
		mockFunc     func(ctx context.Context, params *cloudwatchlogs.DeleteLogGroupInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DeleteLogGroupOutput, error)
		wantErr      bool
	}{
		{
			name:         "successful delete",
			functionName: "test-function",
			mockFunc: func(ctx context.Context, params *cloudwatchlogs.DeleteLogGroupInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DeleteLogGroupOutput, error) {
				expectedLogGroupName := "/aws/lambda/test-function"
				if *params.LogGroupName != expectedLogGroupName {
					t.Errorf("expected log group name %s, got %s", expectedLogGroupName, *params.LogGroupName)
				}
				return &cloudwatchlogs.DeleteLogGroupOutput{}, nil
			},
			wantErr: false,
		},
		{
			name:         "delete error",
			functionName: "test-function",
			mockFunc: func(ctx context.Context, params *cloudwatchlogs.DeleteLogGroupInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DeleteLogGroupOutput, error) {
				return nil, errors.New("InternalError")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockLogsClient{
				deleteLogGroupFunc: tt.mockFunc,
			}
			svc := &Service{client: mock}

			err := svc.DeleteFunctionLogGroup(context.Background(), tt.functionName)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteFunctionLogGroup() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

package client

import (
	"context"
	"os"
	"testing"
)

func TestNewAWSClient(t *testing.T) {
	tests := []struct {
		name    string
		region  string
		profile string
		wantErr bool
	}{
		{
			name:    "create client with default region and profile",
			region:  "",
			profile: "",
			wantErr: false,
		},
		{
			name:    "create client with specific region",
			region:  "us-west-2",
			profile: "",
			wantErr: false,
		},
		{
			name:    "create client with region and profile",
			region:  "ap-northeast-1",
			profile: "default",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			client, err := NewAWSClient(ctx, tt.region, tt.profile)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAWSClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if client == nil {
					t.Error("NewAWSClient() returned nil client")
					return
				}
				if client.Lambda == nil {
					t.Error("NewAWSClient() Lambda client is nil")
				}
				if client.Logs == nil {
					t.Error("NewAWSClient() Logs client is nil")
				}
				if client.CloudFormation == nil {
					t.Error("NewAWSClient() CloudFormation client is nil")
				}
				if tt.region != "" && client.Config.Region != tt.region {
					t.Errorf("NewAWSClient() region = %v, want %v", client.Config.Region, tt.region)
				}
			}
		})
	}
}

func TestGetProxyURL(t *testing.T) {
	tests := []struct {
		name      string
		setEnv    map[string]string
		wantProxy bool
	}{
		{
			name:      "no proxy set",
			setEnv:    map[string]string{},
			wantProxy: false,
		},
		{
			name: "HTTPS_PROXY set",
			setEnv: map[string]string{
				"HTTPS_PROXY": "http://proxy.example.com:8080",
			},
			wantProxy: true,
		},
		{
			name: "HTTP_PROXY set",
			setEnv: map[string]string{
				"HTTP_PROXY": "http://proxy.example.com:3128",
			},
			wantProxy: true,
		},
		{
			name: "https_proxy set (lowercase)",
			setEnv: map[string]string{
				"https_proxy": "http://proxy.example.com:8080",
			},
			wantProxy: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all proxy environment variables
			os.Unsetenv("HTTPS_PROXY")
			os.Unsetenv("https_proxy")
			os.Unsetenv("HTTP_PROXY")
			os.Unsetenv("http_proxy")

			// Set test environment variables
			for k, v := range tt.setEnv {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			proxyURL := getProxyURL()
			if tt.wantProxy && proxyURL == nil {
				t.Error("getProxyURL() expected proxy URL but got nil")
			}
			if !tt.wantProxy && proxyURL != nil {
				t.Errorf("getProxyURL() expected nil but got %v", proxyURL)
			}
		})
	}
}

package client

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/url"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

// AWSClient wraps AWS service clients
type AWSClient struct {
	Lambda         *lambda.Client
	Logs           *cloudwatchlogs.Client
	CloudFormation *cloudformation.Client
	Config         aws.Config
}

// NewAWSClient creates a new AWS client with support for profile and proxy
func NewAWSClient(ctx context.Context, region, profile string) (*AWSClient, error) {
	var opts []func(*config.LoadOptions) error

	// Set region if provided
	if region != "" {
		opts = append(opts, config.WithRegion(region))
	}

	// Set profile if provided
	if profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	}

	// Configure HTTP client with proxy support
	httpClient := &http.Client{
		Transport: createTransportWithProxy(),
	}
	opts = append(opts, config.WithHTTPClient(httpClient))

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return &AWSClient{
		Lambda:         lambda.NewFromConfig(cfg),
		Logs:           cloudwatchlogs.NewFromConfig(cfg),
		CloudFormation: cloudformation.NewFromConfig(cfg),
		Config:         cfg,
	}, nil
}

// createTransportWithProxy creates an HTTP transport with proxy configuration
// Respects HTTP_PROXY, HTTPS_PROXY, and NO_PROXY environment variables
func createTransportWithProxy() *http.Transport {
	transport := http.DefaultTransport.(*http.Transport).Clone()

	// Get proxy settings from environment variables
	proxyURL := getProxyURL()
	if proxyURL != nil {
		transport.Proxy = http.ProxyURL(proxyURL)
	} else {
		// Use default proxy from environment (HTTP_PROXY, HTTPS_PROXY, NO_PROXY)
		transport.Proxy = http.ProxyFromEnvironment
	}

	// Configure TLS settings
	transport.TLSClientConfig = &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	return transport
}

// getProxyURL retrieves proxy URL from environment variables
// Checks HTTPS_PROXY first, then HTTP_PROXY
func getProxyURL() *url.URL {
	// Check HTTPS_PROXY first (case-insensitive)
	proxyEnv := os.Getenv("HTTPS_PROXY")
	if proxyEnv == "" {
		proxyEnv = os.Getenv("https_proxy")
	}

	// Fall back to HTTP_PROXY
	if proxyEnv == "" {
		proxyEnv = os.Getenv("HTTP_PROXY")
		if proxyEnv == "" {
			proxyEnv = os.Getenv("http_proxy")
		}
	}

	if proxyEnv == "" {
		return nil
	}

	proxyURL, err := url.Parse(proxyEnv)
	if err != nil {
		return nil
	}

	return proxyURL
}

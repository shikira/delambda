# clambda

A powerful CLI tool to safely delete AWS Lambda functions with VPC attachments.

## Problem Statement

Deleting VPC-attached Lambda functions can be extremely time-consuming due to:

1. **IPv6 configuration issues** - IPv6 must be disabled before VPC detachment
2. **VPC detachment delays** - Detaching VPCs from Lambda functions is a slow operation
3. **Log group persistence** - CloudWatch Logs groups remain after function deletion, causing recreation failures

`clambda` solves these problems by providing a streamlined, automated workflow for Lambda function deletion.

## Installation

### Using install script (Recommended)

For Linux and macOS:

```bash
# Install the latest version
curl -sSfL https://raw.githubusercontent.com/shikira/clambda/main/install.sh | bash

# Install a specific version
curl -sSfL https://raw.githubusercontent.com/shikira/clambda/main/install.sh | bash -s v1.0.0
```

### Using Go install

Requires Go 1.25 or later:

```bash
go install github.com/shirasu/clambda/cmd/clambda@latest
```

### Download binary manually

Download the appropriate binary for your platform from the [releases page](https://github.com/shikira/clambda/releases) and add it to your PATH.

## Features

- List all Lambda functions with VPC status
- List Lambda functions in a CloudFormation stack
- Disable IPv6 for Lambda functions
- Detach VPCs from Lambda functions (single or all in a stack)
- Delete Lambda functions (single or all in a stack)
- Delete associated CloudWatch Logs log groups
- Comprehensive error handling and progress feedback
- Built with Domain-Driven Design (DDD) architecture

## Usage

### List Lambda functions

```bash
# List all Lambda functions in the account and region
clambda list

# List Lambda functions in a specific CloudFormation stack
clambda list --stack my-stack
```

### Delete Lambda functions

```bash
# Delete a Lambda function and its log group (VPC will be automatically detached if attached)
clambda delete --lambda my-function

# Delete a Lambda function without deleting its log group
clambda delete --lambda my-function --without-logs

# Delete all Lambda functions in a CloudFormation stack (including log groups)
clambda delete --stack my-stack

# Delete all Lambda functions in a stack without deleting log groups
clambda delete --stack my-stack --without-logs
```

### Detach VPC from Lambda functions

```bash
# Detach VPC from a single Lambda function
clambda detach --lambda my-function

# Detach VPC from all Lambda functions in a CloudFormation stack
clambda detach --stack my-stack
```

## Configuration

### AWS Region and Profile

```bash
# Using flags
clambda --region us-west-2 --profile dev list

# Using environment variables
export AWS_REGION=us-west-2
export AWS_PROFILE=dev
clambda list
```

### Proxy Support

`clambda` automatically respects standard HTTP proxy environment variables:
- `HTTP_PROXY` / `http_proxy`
- `HTTPS_PROXY` / `https_proxy`
- `NO_PROXY` / `no_proxy`

## Development

### Prerequisites for Development

- Go 1.25 or later
- AWS credentials configured
- Node.js and npm (for CDK)

### Available Make Commands

```bash
make help         # Show all available commands
make install      # Set up development environment
make build        # Build the tool and CDK project
make test         # Run unit tests
make test-integ   # Run integration tests
make lint         # Run linters
make fmt          # Format code
make deploy       # Deploy CDK test infrastructure
make destroy      # Destroy CDK test infrastructure
```

### Build from source

```bash
make build
```

### Deploy Test Environment

```bash
# Deploy with default AWS credentials
make deploy

# Deploy with a specific profile
make deploy PROFILE=dev
```

## Troubleshooting

### Access denied errors

Ensure your AWS credentials have the following permissions:
- `lambda:ListFunctions`
- `lambda:GetFunction`
- `lambda:UpdateFunctionConfiguration`
- `lambda:DeleteFunction`
- `logs:DescribeLogGroups`
- `logs:DeleteLogGroup`
- `cloudformation:DescribeStacks`
- `cloudformation:ListStackResources`

## License

MIT License

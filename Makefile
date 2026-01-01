.PHONY: help install clean build test test-integ test-integ-full lint fmt deploy destroy

# Variables
PROFILE ?=
REGION ?= us-east-1

# Set profile flag only if PROFILE is specified
ifdef PROFILE
PROFILE_FLAG = --profile $(PROFILE)
else
PROFILE_FLAG =
endif

# Default target
help:
	@echo "Available targets:"
	@echo "  install      - Set up development environment"
	@echo "  clean        - Remove build artifacts"
	@echo "  build        - Build the tool and CDK project"
	@echo "  test         - Run unit tests for tool and CDK"
	@echo "  test-integ   - Run integration tests (read-only)"
	@echo "  test-integ-full - Run full integration tests (includes destructive operations on test stack)"
	@echo "  lint         - Run linters"
	@echo "  fmt          - Format code"
	@echo "  deploy       - Deploy CDK test infrastructure"
	@echo "  destroy      - Destroy CDK test infrastructure"
	@echo ""
	@echo "Variables:"
	@echo "  PROFILE      - AWS profile to use (optional, omit to use default AWS credentials)"
	@echo "  REGION       - AWS region to use (default: us-east-1)"

# Set up development environment
install:
	@echo "Installing Go dependencies..."
	go mod download
	@echo "Installing CDK dependencies..."
	cd cdk && pnpm install
	@echo "Development environment setup complete!"

# Remove build artifacts
clean:
	@echo "Cleaning Go build artifacts..."
	rm -f delambda
	rm -f coverage.out coverage.html
	rm -f *.test
	@echo "Cleaning CDK build artifacts..."
	cd cdk && rm -rf cdk.out .cdk.staging
	cd cdk && rm -f bin/*.js bin/*.js.map bin/*.d.ts
	cd cdk && rm -f lib/*.js lib/*.js.map lib/*.d.ts
	@echo "Clean complete!"

# Build the tool and CDK project
build:
	@echo "Building delambda tool..."
	go build -o delambda ./cmd/delambda
	@echo "Building CDK project..."
	cd cdk && pnpm run build
	@echo "Build complete!"

# Run unit tests
test:
	@echo "Running Go tests..."
	go test ./... -v -coverprofile=coverage.out
	@echo "Running CDK tests..."
	cd cdk && pnpm test
	@echo "Tests complete!"

# Run integration tests
test-integ:
ifdef PROFILE
	@echo "Running integration tests (profile: $(PROFILE), region: $(REGION))..."
	@echo ""
	@echo "=== Test 1: List all Lambda functions ==="
	./delambda list --profile $(PROFILE) --region $(REGION)
	@echo ""
	@echo "=== Test 2: List Lambda functions in CDK stack ==="
	./delambda list --stack TestInfrastructureStack --profile $(PROFILE) --region $(REGION) || echo "Warning: TestInfrastructureStack not found (run 'make deploy' first)"
	@echo ""
	@echo "=== Test 3: Show help message ==="
	./delambda help
	@echo ""
	@echo "=== Test 4: Validate detach command arguments ==="
	./delambda detach 2>&1 | grep -q "Either --lambda or --stack must be specified" && echo "✓ Detach validation works" || echo "✗ Detach validation failed"
	@echo ""
	@echo "=== Test 5: Validate delete command arguments ==="
	./delambda delete 2>&1 | grep -q "Either --lambda or --stack must be specified" && echo "✓ Delete validation works" || echo "✗ Delete validation failed"
	@echo ""
	@echo "=== Test 6: Validate delete-logs command arguments ==="
	./delambda delete-logs 2>&1 | grep -q "Log group name is required" && echo "✓ Delete-logs validation works" || echo "✗ Delete-logs validation failed"
else
	@echo "Running integration tests (region: $(REGION))..."
	@echo ""
	@echo "=== Test 1: List all Lambda functions ==="
	./delambda list --region $(REGION)
	@echo ""
	@echo "=== Test 2: List Lambda functions in CDK stack ==="
	./delambda list --stack TestInfrastructureStack --region $(REGION) || echo "Warning: TestInfrastructureStack not found (run 'make deploy' first)"
	@echo ""
	@echo "=== Test 3: Show help message ==="
	./delambda help
	@echo ""
	@echo "=== Test 4: Validate detach command arguments ==="
	./delambda detach 2>&1 | grep -q "Either --lambda or --stack must be specified" && echo "✓ Detach validation works" || echo "✗ Detach validation failed"
	@echo ""
	@echo "=== Test 5: Validate delete command arguments ==="
	./delambda delete 2>&1 | grep -q "Either --lambda or --stack must be specified" && echo "✓ Delete validation works" || echo "✗ Delete validation failed"
	@echo ""
	@echo "=== Test 6: Validate delete-logs command arguments ==="
	./delambda delete-logs 2>&1 | grep -q "Log group name is required" && echo "✓ Delete-logs validation works" || echo "✗ Delete-logs validation failed"
endif
	@echo ""
	@echo "=== Integration tests complete! ==="

# Run full integration tests (includes destructive operations)
test-integ-full: test-integ
ifdef PROFILE
	@echo ""
	@echo "=== Running full integration tests with destructive operations ==="
	@echo "Warning: This will modify the TestInfrastructureStack!"
	@echo ""
	@echo "=== Test 7: Detach VPC from stack (dry-run - list first) ==="
	./delambda list --stack TestInfrastructureStack --profile $(PROFILE) --region $(REGION) || (echo "Error: TestInfrastructureStack not found. Run 'make deploy' first." && exit 1)
	@echo ""
	@echo "=== Test 8: Actually detach VPC from test stack ==="
	@read -p "Press Enter to detach VPC from TestInfrastructureStack (Ctrl+C to cancel)..." confirm
	./delambda detach --stack TestInfrastructureStack --profile $(PROFILE) --region $(REGION)
	@echo ""
	@echo "=== Test 9: Verify VPC was detached ==="
	./delambda list --stack TestInfrastructureStack --profile $(PROFILE) --region $(REGION)
	@echo ""
	@echo "Note: Re-run 'make deploy' to restore the test infrastructure with VPC attachments"
else
	@echo ""
	@echo "=== Running full integration tests with destructive operations ==="
	@echo "Warning: This will modify the TestInfrastructureStack!"
	@echo ""
	@echo "=== Test 7: Detach VPC from stack (dry-run - list first) ==="
	./delambda list --stack TestInfrastructureStack --region $(REGION) || (echo "Error: TestInfrastructureStack not found. Run 'make deploy' first." && exit 1)
	@echo ""
	@echo "=== Test 8: Actually detach VPC from test stack ==="
	@read -p "Press Enter to detach VPC from TestInfrastructureStack (Ctrl+C to cancel)..." confirm
	./delambda detach --stack TestInfrastructureStack --region $(REGION)
	@echo ""
	@echo "=== Test 9: Verify VPC was detached ==="
	./delambda list --stack TestInfrastructureStack --region $(REGION)
	@echo ""
	@echo "Note: Re-run 'make deploy' to restore the test infrastructure with VPC attachments"
endif
	@echo ""
	@echo "=== Full integration tests complete! ==="

# Run linters
lint:
	@echo "Running Go linters..."
	go vet ./...
	@echo "Running Go fmt check..."
	@test -z "$$(gofmt -l $$(find . -name '*.go' -not -path './cdk/*'))" || (echo "Code is not formatted. Run 'make fmt'" && exit 1)
	@echo "Lint complete!"

# Format code
fmt:
	@echo "Formatting Go code..."
	go fmt ./...
	@echo "Format complete!"

# Deploy CDK test infrastructure
deploy:
ifdef PROFILE
	@echo "Deploying CDK test infrastructure (profile: $(PROFILE))..."
else
	@echo "Deploying CDK test infrastructure (using default AWS credentials)..."
endif
	cd cdk && pnpm run build
	cd cdk && pnpm run cdk deploy $(PROFILE_FLAG) --require-approval never
	@echo "Deploy complete!"

# Destroy CDK test infrastructure
destroy:
ifdef PROFILE
	@echo "Destroying CDK test infrastructure (profile: $(PROFILE))..."
else
	@echo "Destroying CDK test infrastructure (using default AWS credentials)..."
endif
	cd cdk && pnpm run cdk destroy $(PROFILE_FLAG) --force
	@echo "Destroy complete!"

# Generate coverage report
coverage:
	@echo "Generating coverage report..."
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

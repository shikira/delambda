.PHONY: help install clean build test test-integ lint fmt deploy destroy

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
	@echo "  test-integ   - Run integration tests"
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
	rm -f clambda
	rm -f coverage.out coverage.html
	rm -f *.test
	@echo "Cleaning CDK build artifacts..."
	cd cdk && rm -rf cdk.out .cdk.staging
	cd cdk && rm -f bin/*.js bin/*.js.map bin/*.d.ts
	cd cdk && rm -f lib/*.js lib/*.js.map lib/*.d.ts
	@echo "Clean complete!"

# Build the tool and CDK project
build:
	@echo "Building clambda tool..."
	go build -o clambda ./cmd/clambda
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
	./clambda --profile $(PROFILE) --region $(REGION) list
else
	@echo "Running integration tests (region: $(REGION))..."
	./clambda --region $(REGION) list
endif
	@echo "Integration tests complete!"

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
	cd cdk && npx cdk deploy $(PROFILE_FLAG) --require-approval never
	@echo "Deploy complete!"

# Destroy CDK test infrastructure
destroy:
ifdef PROFILE
	@echo "Destroying CDK test infrastructure (profile: $(PROFILE))..."
else
	@echo "Destroying CDK test infrastructure (using default AWS credentials)..."
endif
	cd cdk && npx cdk destroy $(PROFILE_FLAG) --force
	@echo "Destroy complete!"

# Generate coverage report
coverage:
	@echo "Generating coverage report..."
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# A2A Serverless Go Makefile

.PHONY: test build clean deploy help

# Default target
help:
	@echo "Available targets:"
	@echo "  test     - Run all tests"
	@echo "  build    - Build Lambda binary"
	@echo "  clean    - Clean build artifacts"
	@echo "  deploy   - Create deployment package"
	@echo "  help     - Show this help message"

# Run tests
test:
	go test ./...

# Build for Lambda (Linux AMD64)
build:
	GOOS=linux GOARCH=amd64 go build -o bootstrap cmd/lambda/main.go

# Clean build artifacts
clean:
	rm -f bootstrap lambda-deployment.zip

# Create deployment package
deploy: build
	zip lambda-deployment.zip bootstrap
	@echo "Deployment package created: lambda-deployment.zip"

# Development build (current platform)
dev-build:
	go build -o a2a-serverless cmd/lambda/main.go
# Task Learnings

This file contains detailed learnings from each task to help agents be more efficient and correct in future implementations.

## Task 1: Set up project structure and core interfaces

### A2A SDK Integration (CRITICAL)
- **ALWAYS use official A2A SDK types** - Don't create custom types that duplicate SDK functionality
- The SDK provides: `a2a.Task`, `a2a.Message`, `a2a.Event`, `a2a.AgentCard`, etc.
- The SDK provides: `a2asrv.RequestHandler` interface for server implementations
- Only create serverless-specific types for infrastructure concerns (storage metadata, cloud config)
- Check SDK documentation first: explore the installed package at `/Users/chris/go/pkg/mod/github.com/a2aproject/a2a-go@<version>/`

### AWS SDK Integration (CRITICAL)
- **ALWAYS use official AWS SDK v2 types and interfaces** - Don't create custom types that duplicate AWS SDK functionality
- The SDK provides: DynamoDB types (`dynamodb.Client`, `types.AttributeValue*`, etc.)
- The SDK provides: SQS types (`sqs.Client`, `sqs.SendMessageInput`, etc.)
- The SDK provides: Lambda types (`lambda.Client`, context types, etc.)
- Use AWS SDK v2 (not v1 which is deprecated): `github.com/aws/aws-sdk-go-v2/*`
- Only create wrapper types for business logic abstraction, not AWS service duplication
- Required packages: `aws-sdk-go-v2/config`, `aws-sdk-go-v2/service/dynamodb`, `aws-sdk-go-v2/service/sqs`

### Project Structure Lessons
- Keep the project structure flat and simple: `cmd/lambda`, `internal/a2a`, `internal/handler`
- Separate concerns cleanly:
  - `internal/a2a/`: A2A protocol implementation and serverless-specific types
  - `internal/handler/`: HTTP/Lambda request routing
  - `cmd/lambda/`: AWS Lambda entry point and initialization
- Use interfaces for storage (`TaskStore`, `EventStore`, `PushNotifier`) to enable testing and multiple implementations

### Go Module Management
- The A2A SDK requires Go 1.24.4+ (it will auto-upgrade the Go version)
- Use `go mod tidy` frequently to keep dependencies clean
- Make dependencies explicit in go.mod (not indirect)

### Testing Strategy
- Test serverless-specific functionality, not SDK functionality
- Focus tests on: validation functions, storage metadata, configuration parsing
- Use the SDK types in tests to ensure compatibility
- Keep tests simple and focused (grug-brain principle)

### Development Workflow
- Start with types and interfaces first
- Implement storage interfaces with AWS-specific implementations
- Create the A2A handler implementing `a2asrv.RequestHandler`
- Wire everything together in the Lambda main function
- Test incrementally: `go test ./...` and `make dev-build` frequently

### Common Pitfalls to Avoid
- Don't reinvent A2A protocol types - use the SDK
- Don't forget to import `time` package when using `time.Now()`
- Don't use `interface{}` for A2A events - use proper SDK event types
- Don't make AWS SDK dependencies indirect - they should be explicit
- Don't skip environment variable validation in production code

### Environment Configuration Pattern
```go
func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
```

### Interface Implementation Verification Pattern
```go
// Verify that ServerlessA2AHandler implements the RequestHandler interface
var _ a2asrv.RequestHandler = (*ServerlessA2AHandler)(nil)
```

This ensures compile-time verification that interfaces are properly implemented.

### File Organization Established
- `internal/a2a/types.go`: Serverless-specific configuration and storage types only
- `internal/a2a/server.go`: A2A RequestHandler implementation for serverless
- `internal/a2a/aws_storage.go`: AWS-specific storage implementations (DynamoDB, SQS)
- `internal/handler/handler.go`: HTTP to JSON-RPC routing and agent card serving
- `cmd/lambda/main.go`: AWS Lambda initialization and environment setup

## Task 2: Implement configuration management with cloud provider abstraction

### Configuration Management Learnings
- Environment-based configuration is essential for serverless deployments
- Create a `ConfigLoader` struct to centralize configuration loading logic
- Validate configuration at multiple levels: environment variables, provider configs, and complete configuration
- Use clear, descriptive error messages for configuration validation failures

### Cloud Provider Abstraction Patterns
- Define interfaces for cloud provider operations (`CloudProviderInterface`)
- Implement provider-specific structs that satisfy the interface (`AWSProvider`, `GCPProvider`, `LocalProvider`)
- Each provider should handle its own validation and configuration retrieval
- Use factory pattern for provider creation based on configuration

### A2A SDK Type Integration Discoveries
- `a2a.AgentCapabilities` is a struct with specific fields, not a string slice
- Fields include: `Extensions`, `PushNotifications`, `StateTransitionHistory`, `Streaming`
- Use pointer fields for optional boolean capabilities (`*bool`)
- Only set capability pointers when environment variables are explicitly provided

### Environment Variable Handling Best Practices
- Create helper functions for different data types: `getEnvOrDefault`, `getEnvOrDefaultInt`, `getEnvOrDefaultBool`
- Validate required environment variables upfront with `ValidateEnvironmentVariables()`
- Use provider-specific validation for provider-specific requirements
- Provide sensible defaults where appropriate

### Testing Configuration Systems
- Test all configuration loading scenarios: valid configs, missing variables, invalid values
- Test each cloud provider separately with comprehensive validation scenarios
- Use environment variable cleanup in tests to ensure isolation
- Create helper functions for test setup and teardown (`clearTestEnv()`)
- Test both positive and negative cases with clear error message validation

### Configuration Structure Patterns
```go
// Separate configuration concerns
type ServerlessConfig struct {
    AgentID     string
    AgentCard   a2a.AgentCard
    CloudConfig CloudProviderConfig
    LogLevel    string
}

// Provider-specific configuration
type CloudProviderConfig struct {
    Provider string
    AWS      *AWSConfig
    // Future providers...
}
```

### Boolean Pointer Handling for A2A Capabilities
```go
// Only set pointers when explicitly configured
if os.Getenv("A2A_AGENT_PUSH_NOTIFICATIONS") != "" {
    pushNotifications := getEnvOrDefaultBool("A2A_AGENT_PUSH_NOTIFICATIONS", false)
    capabilities.PushNotifications = &pushNotifications
}
```

### Test Helper Patterns
```go
// Helper for boolean pointers in tests
func boolPtr(b bool) *bool {
    return &b
}

// Helper for capability comparison
func compareCapabilities(a, b a2a.AgentCapabilities) bool {
    // Compare each field including nil checks
}
```

### Configuration Validation Error Patterns
- Provide specific error messages that identify the missing or invalid field
- Include context about which provider or configuration section has the issue
- Use wrapped errors with `fmt.Errorf("context: %w", err)` for error chains
- Validate at the appropriate level (environment, provider, complete config)

### Multi-Provider Support Strategy
- Design interfaces that can accommodate different cloud providers
- Implement local provider for development and testing
- Prepare GCP provider interface even if not fully implemented
- Use consistent patterns across all provider implementations

### Configuration Loading Workflow
1. Load basic A2A configuration (agent ID, card details)
2. Load cloud provider configuration based on CLOUD_PROVIDER env var
3. Create provider instance and validate provider-specific config
4. Combine into complete ServerlessConfig and validate holistically
5. Return validated configuration ready for use

### Key Success Metrics for Configuration Tasks
- All tests pass with comprehensive coverage
- Configuration loading works for all supported providers
- Clear error messages for all failure scenarios
- Environment variable validation catches missing requirements
- Provider abstraction enables easy addition of new providers
- Configuration system integrates properly with existing A2A and AWS SDK types
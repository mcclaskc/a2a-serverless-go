# A2A Serverless Go Implementation

This is a Go implementation of the A2A (Agent-to-Agent) protocol for serverless environments, specifically designed for AWS Lambda.

## Project Structure

```
a2a-serverless-go/
├── cmd/
│   └── lambda/           # Lambda entry point
│       └── main.go
├── internal/
│   ├── a2a/             # A2A protocol types and utilities
│   │   ├── types.go
│   │   └── types_test.go
│   └── handler/         # HTTP request handlers
│       └── handler.go
├── go.mod
└── README.md
```

## Dependencies

- **a2a-go SDK**: Official A2A Go SDK for protocol implementation
- **AWS Lambda Go**: AWS Lambda runtime for Go
- **AWS SDK Go**: AWS services integration (DynamoDB, SQS)

## Development

### Running Tests

```bash
go test ./...
```

### Building for Lambda

```bash
GOOS=linux GOARCH=amd64 go build -o bootstrap cmd/lambda/main.go
zip lambda-deployment.zip bootstrap
```

## Architecture

This implementation follows the grug-brain development philosophy:
- Simple, procedural code over complex abstractions
- Pure functions where possible
- Clear separation of concerns
- Comprehensive test coverage

## Core Components

### A2A Integration (`internal/a2a/`)

- **Official A2A SDK**: Uses `github.com/a2aproject/a2a-go` for all protocol types
- **Serverless Types**: `ServerlessConfig`, `TaskStorage`, `EventStorage` for serverless-specific needs
- **Server Implementation**: `ServerlessA2AHandler` implements the official `RequestHandler` interface
- **AWS Storage**: DynamoDB-based implementations for `TaskStore` and `EventStore`
- **Push Notifications**: SQS-based push notification system

### Handler (`internal/handler/handler.go`)

- HTTP to JSON-RPC request routing
- Agent card serving (GET /)
- A2A protocol method handling (tasks/get, tasks/cancel, message/send)
- CORS support for web clients

### Lambda Entry Point (`cmd/lambda/main.go`)

- AWS Lambda integration with API Gateway
- AWS SDK initialization (DynamoDB, SQS)
- Environment-based configuration
- A2A handler setup with official SDK integration

## Configuration

The Lambda function uses environment variables for configuration:

- `AGENT_ID`: Unique identifier for the agent (default: "serverless-agent-1")
- `AGENT_NAME`: Human-readable agent name (default: "A2A Serverless Agent")
- `AGENT_URL`: Public URL where the agent is accessible
- `DYNAMODB_TABLE`: DynamoDB table for task storage (default: "a2a-tasks")
- `DYNAMODB_EVENTS_TABLE`: DynamoDB table for event storage (default: "a2a-events")
- `SQS_QUEUE_URL`: SQS queue URL for push notifications
- `LOG_LEVEL`: Logging level (default: "info")

## Testing

All core functionality is covered by unit tests following grug-brain principles:
- Simple, focused test cases using official A2A SDK types
- Clear test names and assertions
- Comprehensive coverage of serverless-specific functionality
- Fast execution without external dependencies
# a2a-serverless-go

This AGENTS.md file is a live document and changes as the project progresses. Check it frequently for updates. Also make your own updates but ask permission first.

## Task Workflow

### Starting a New Task
1. **Always review AGENTS.md** - Read this file for current project guidelines and critical requirements
2. **Always review TASK_LEARNINGS.md** - Read previous task learnings to avoid repeating mistakes and leverage successful patterns

### Ending a Task
1. **Always add task learnings to TASK_LEARNINGS.md** - Document specific learnings, patterns, pitfalls, and solutions discovered during the task to help future agents be more efficient and correct

## Project Overview
a2a-serverless is a module that allows implementation of a2a servers and clients in serverless environments. Read the README.md to understand what the humans do.

## Development Philosophy
- **Test-Driven Development**: Write failing tests first, then implement, then repeat until all tests are green
- **Incremental Implementation**: Do things one test/function at a time instead of implementing everything at once
- **Grug-Brain Approach**: Keep it simple - complexity is bad (https://grugbrain.dev/)
- **Procedural over OO/Functional**: Prefer procedural programming patterns in Go

## Code Style Guidelines
- Functions should be pure whenever possible
- Side effects should be isolated within single functions
- When multiple side effects are required, manage them in a clear 'orchestrator' function
- Comments should explain 'why' not 'how' or 'what'
- Keep the project structure flat and simple

## Architecture Principles
- Use the official A2A SDK: https://github.com/a2aproject/a2a-go
- Use go-service/cmd/lambda to wrap Lambda handler code
- Separate concerns cleanly:
  - `internal/a2a/`: A2A protocol implementation and serverless-specific types
  - `internal/handler/`: HTTP/Lambda request routing
  - `cmd/lambda/`: AWS Lambda entry point and initialization

## Critical Requirements

### A2A SDK Integration (CRITICAL)
- **ALWAYS use official A2A SDK types** - Don't create custom types that duplicate SDK functionality
- The SDK provides: `a2a.Task`, `a2a.Message`, `a2a.Event`, `a2a.AgentCard`, `a2asrv.RequestHandler`, etc.
- Only create serverless-specific types for infrastructure concerns (storage metadata, cloud config)
- Verify interface compliance at compile time: `var _ a2asrv.RequestHandler = (*YourHandler)(nil)`

### AWS SDK Integration (CRITICAL)
- **ALWAYS use official AWS SDK v2 types and interfaces** - Don't create custom types that duplicate AWS SDK functionality
- Use AWS SDK v2 (not deprecated v1): `github.com/aws/aws-sdk-go-v2/*`
- The SDK provides: `dynamodb.Client`, `sqs.Client`, `types.AttributeValue*`, etc.
- Only create wrapper types for business logic abstraction, not AWS service duplication
- Required packages: `aws-sdk-go-v2/config`, `aws-sdk-go-v2/service/dynamodb`, `aws-sdk-go-v2/service/sqs`

## Current Project Status
- ✅ Task 1: Project structure and core interfaces established
- ✅ Task 2: Configuration management with cloud provider abstraction implemented
- Foundation is solid with proper A2A and AWS SDK integration
- All tests pass and build succeeds
- Ready for business logic and A2A method implementations

## Key Files
- `TASK_LEARNINGS.md`: Detailed learnings from each completed task
- `internal/a2a/types.go`: Serverless-specific configuration and storage types
- `internal/a2a/config.go`: Configuration management and cloud provider abstraction
- `internal/a2a/server.go`: A2A RequestHandler implementation for serverless
- `internal/a2a/aws_storage.go`: AWS-specific storage implementations
- `internal/handler/handler.go`: HTTP to JSON-RPC routing and agent card serving
- `cmd/lambda/main.go`: AWS Lambda initialization and environment setup

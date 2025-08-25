package main

import (
	"fmt"
	"log"
	"os"

	"github.com/a2aproject/a2a-serverless/internal/a2a"
)

func main() {
	// Example of using the configuration system
	
	// Set up example environment variables
	os.Setenv("A2A_AGENT_ID", "example-agent-123")
	os.Setenv("A2A_AGENT_NAME", "Example Agent")
	os.Setenv("A2A_AGENT_URL", "https://example-agent.com")
	os.Setenv("A2A_AGENT_DESCRIPTION", "An example A2A agent")
	os.Setenv("A2A_AGENT_VERSION", "1.0.0")
	os.Setenv("A2A_AGENT_PUSH_NOTIFICATIONS", "true")
	os.Setenv("A2A_AGENT_STATE_HISTORY", "false")
	os.Setenv("A2A_LOG_LEVEL", "info")
	
	// Example 1: Local provider
	fmt.Println("=== Local Provider Example ===")
	os.Setenv("CLOUD_PROVIDER", "local")
	os.Setenv("LOCAL_STORAGE_PATH", "./example_storage")
	os.Setenv("LOCAL_EVENT_PATH", "./example_events")
	
	loader := a2a.NewConfigLoader()
	
	// Load complete serverless configuration
	config, err := loader.LoadServerlessConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	
	fmt.Printf("Agent ID: %s\n", config.AgentID)
	fmt.Printf("Agent Name: %s\n", config.AgentCard.Name)
	fmt.Printf("Agent URL: %s\n", config.AgentCard.URL)
	fmt.Printf("Cloud Provider: %s\n", config.CloudConfig.Provider)
	fmt.Printf("Log Level: %s\n", config.LogLevel)
	
	// Create cloud provider instance
	provider, err := loader.CreateCloudProvider(config.CloudConfig)
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}
	
	fmt.Printf("Provider Type: %s\n", provider.GetProviderType())
	fmt.Printf("Storage Config: %+v\n", provider.GetStorageConfig())
	fmt.Printf("Event Config: %+v\n", provider.GetEventConfig())
	
	// Example 2: AWS provider
	fmt.Println("\n=== AWS Provider Example ===")
	os.Setenv("CLOUD_PROVIDER", "aws")
	os.Setenv("AWS_REGION", "us-east-1")
	os.Setenv("AWS_SQS_QUEUE_URL", "https://sqs.us-east-1.amazonaws.com/123456789/example-queue")
	os.Setenv("AWS_DYNAMODB_TABLE", "example-table")
	
	// Load AWS configuration
	awsConfig, err := loader.LoadServerlessConfig()
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}
	
	fmt.Printf("Cloud Provider: %s\n", awsConfig.CloudConfig.Provider)
	if awsConfig.CloudConfig.AWS != nil {
		fmt.Printf("AWS Region: %s\n", awsConfig.CloudConfig.AWS.Region)
		fmt.Printf("AWS SQS Queue: %s\n", awsConfig.CloudConfig.AWS.SQSQueueURL)
		fmt.Printf("AWS DynamoDB Table: %s\n", awsConfig.CloudConfig.AWS.DynamoDBTable)
	}
	
	// Create AWS provider instance
	awsProvider, err := loader.CreateCloudProvider(awsConfig.CloudConfig)
	if err != nil {
		log.Fatalf("Failed to create AWS provider: %v", err)
	}
	
	fmt.Printf("Provider Type: %s\n", awsProvider.GetProviderType())
	fmt.Printf("Storage Config: %+v\n", awsProvider.GetStorageConfig())
	fmt.Printf("Event Config: %+v\n", awsProvider.GetEventConfig())
	
	// Example 3: Environment validation
	fmt.Println("\n=== Environment Validation Example ===")
	
	// Clear required variables to test validation
	os.Unsetenv("A2A_AGENT_ID")
	
	err = a2a.ValidateEnvironmentVariables()
	if err != nil {
		fmt.Printf("Validation error (expected): %v\n", err)
	}
	
	// Restore required variable
	os.Setenv("A2A_AGENT_ID", "example-agent-123")
	
	err = a2a.ValidateEnvironmentVariables()
	if err != nil {
		fmt.Printf("Unexpected validation error: %v\n", err)
	} else {
		fmt.Println("Environment validation passed!")
	}
	
	fmt.Println("\n=== Configuration Example Complete ===")
}
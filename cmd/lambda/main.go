package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	"github.com/a2aproject/a2a-go/a2a"
	"github.com/a2aproject/a2a-serverless/internal/handler"
	a2aTypes "github.com/a2aproject/a2a-serverless/internal/a2a"
)

var h *handler.Handler

func init() {
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	// Create AWS clients
	dynamoClient := dynamodb.NewFromConfig(cfg)
	sqsClient := sqs.NewFromConfig(cfg)

	// Get configuration from environment variables
	tableName := getEnvOrDefault("DYNAMODB_TABLE", "a2a-tasks")
	eventsTable := getEnvOrDefault("DYNAMODB_EVENTS_TABLE", "a2a-events")
	sqsQueueURL := getEnvOrDefault("SQS_QUEUE_URL", "")
	agentName := getEnvOrDefault("AGENT_NAME", "A2A Serverless Agent")
	agentURL := getEnvOrDefault("AGENT_URL", "https://example.com/agent")

	// Create storage implementations
	taskStore := a2aTypes.NewAWSTaskStore(dynamoClient, tableName)
	eventStore := a2aTypes.NewAWSEventStore(dynamoClient, eventsTable)
	pushNotifier := a2aTypes.NewAWSSQSPushNotifier(sqsClient, sqsQueueURL)

	// Create agent card
	agentCard := a2a.AgentCard{
		Name:               agentName,
		URL:                agentURL,
		Description:        "A serverless A2A agent running on AWS Lambda",
		ProtocolVersion:    "1.0",
		Version:            "1.0.0",
		PreferredTransport: a2a.TransportProtocolJSONRPC,
		Capabilities: a2a.AgentCapabilities{
			Streaming:         &[]bool{false}[0], // Non-streaming for serverless
			PushNotifications: &[]bool{true}[0],  // Support push notifications
		},
		Skills: []a2a.AgentSkill{
			{
				ID:          "general",
				Name:        "General Assistant",
				Description: "General purpose AI assistant capabilities",
				Examples:    []string{"Answer questions", "Help with tasks"},
				Tags:        []string{"assistant", "general"},
			},
		},
	}

	// Create serverless config
	serverlessConfig := a2aTypes.ServerlessConfig{
		AgentID:   getEnvOrDefault("AGENT_ID", "serverless-agent-1"),
		AgentCard: agentCard,
		CloudConfig: a2aTypes.CloudProviderConfig{
			Provider: "aws",
			AWS: &a2aTypes.AWSConfig{
				Region:        cfg.Region,
				SQSQueueURL:   sqsQueueURL,
				DynamoDBTable: tableName,
			},
		},
		LogLevel: getEnvOrDefault("LOG_LEVEL", "info"),
	}

	// Create A2A handler
	a2aHandler := a2aTypes.NewServerlessA2AHandler(serverlessConfig, taskStore, eventStore, pushNotifier)

	// Create HTTP handler
	h = handler.NewHandler(a2aHandler, agentCard)
}

func handleLambda(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Convert Lambda request to internal format
	req := handler.Request{
		Method:  request.HTTPMethod,
		URL:     request.Path,
		Headers: request.Headers,
		Body:    request.Body,
	}

	// Process request using A2A handler
	response := h.HandleRequest(req)

	// Convert to Lambda response format
	return events.APIGatewayProxyResponse{
		StatusCode: response.Status,
		Headers:    response.Headers,
		Body:       response.Body,
	}, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	lambda.Start(handleLambda)
}
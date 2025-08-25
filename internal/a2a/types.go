package a2a

import (
	"encoding/json"
	"fmt"
	"time"

	// Import the official A2A SDK types
	"github.com/a2aproject/a2a-go/a2a"
)

// ServerlessConfig holds configuration for A2A serverless operations
type ServerlessConfig struct {
	AgentID     string                   `json:"agent_id"`
	AgentCard   a2a.AgentCard           `json:"agent_card"`
	CloudConfig CloudProviderConfig     `json:"cloud_config"`
	LogLevel    string                  `json:"log_level"`
}

// AWSConfig holds AWS service configuration
type AWSConfig struct {
	SQSQueueURL     string `json:"sqs_queue_url"`
	DynamoDBTable   string `json:"dynamodb_table"`
	Region          string `json:"region"`
	AccessKeyID     string `json:"access_key_id,omitempty"`
	SecretAccessKey string `json:"secret_access_key,omitempty"`
}

// CloudProviderConfig holds configuration for different cloud providers
type CloudProviderConfig struct {
	Provider string     `json:"provider"` // "aws", "gcp", "local"
	AWS      *AWSConfig `json:"aws,omitempty"`
	// Future: GCP, Azure configs can be added here
}

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"` // Always "2.0"
	Method  string      `json:"method"`  // A2A method name
	Params  interface{} `json:"params"`  // Method parameters
	ID      interface{} `json:"id"`      // Request ID
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`           // Always "2.0"
	Result  interface{}   `json:"result,omitempty"`  // Success result
	Error   *JSONRPCError `json:"error,omitempty"`   // Error details
	ID      interface{}   `json:"id"`                // Request ID
}

// JSONRPCError represents a JSON-RPC 2.0 error
type JSONRPCError struct {
	Code    int         `json:"code"`              // Error code
	Message string      `json:"message"`           // Error message
	Data    interface{} `json:"data,omitempty"`    // Additional error data
}

// TaskStorage represents serverless-specific task storage metadata
type TaskStorage struct {
	TaskID       a2a.TaskID        `json:"task_id"`
	ContextID    string            `json:"context_id"`
	StorageKey   string            `json:"storage_key"`
	LastModified int64             `json:"last_modified"`
	TTL          *int64            `json:"ttl,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// EventStorage represents serverless-specific event storage metadata
type EventStorage struct {
	EventID      string            `json:"event_id"`
	TaskID       a2a.TaskID        `json:"task_id"`
	EventType    string            `json:"event_type"`
	StorageKey   string            `json:"storage_key"`
	Timestamp    int64             `json:"timestamp"`
	Processed    bool              `json:"processed"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// ValidateServerlessConfig validates serverless configuration
func ValidateServerlessConfig(config ServerlessConfig) error {
	if config.AgentID == "" {
		return fmt.Errorf("agent_id is required")
	}
	if config.AgentCard.Name == "" {
		return fmt.Errorf("agent_card.name is required")
	}
	if config.AgentCard.URL == "" {
		return fmt.Errorf("agent_card.url is required")
	}
	return ValidateCloudProviderConfig(config.CloudConfig)
}

// ValidateCloudProviderConfig validates cloud provider configuration
func ValidateCloudProviderConfig(config CloudProviderConfig) error {
	if config.Provider == "" {
		return fmt.Errorf("provider is required")
	}
	
	switch config.Provider {
	case "aws":
		if config.AWS == nil {
			return fmt.Errorf("aws configuration is required when provider is 'aws'")
		}
		return ValidateAWSConfig(*config.AWS)
	case "local":
		// Local provider doesn't need additional validation
		return nil
	default:
		return fmt.Errorf("unsupported provider: %s", config.Provider)
	}
}

// ValidateAWSConfig validates AWS configuration
func ValidateAWSConfig(config AWSConfig) error {
	if config.Region == "" {
		return fmt.Errorf("region is required")
	}
	if config.SQSQueueURL == "" {
		return fmt.Errorf("sqs_queue_url is required")
	}
	if config.DynamoDBTable == "" {
		return fmt.Errorf("dynamodb_table is required")
	}
	return nil
}

// ValidateJSONRPCRequest validates a JSON-RPC request
func ValidateJSONRPCRequest(req JSONRPCRequest) error {
	if req.JSONRPC != "2.0" {
		return fmt.Errorf("jsonrpc must be '2.0'")
	}
	if req.Method == "" {
		return fmt.Errorf("method is required")
	}
	if req.ID == nil {
		return fmt.Errorf("id is required")
	}
	return nil
}

// ToJSON serializes any struct to JSON bytes
func ToJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// FromJSON deserializes JSON bytes to a struct
func FromJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// NewTaskStorage creates a new task storage metadata entry
func NewTaskStorage(taskID a2a.TaskID, contextID string) TaskStorage {
	now := time.Now().Unix()
	return TaskStorage{
		TaskID:       taskID,
		ContextID:    contextID,
		StorageKey:   fmt.Sprintf("task_%s", taskID),
		LastModified: now,
		Metadata:     make(map[string]string),
	}
}

// NewEventStorage creates a new event storage metadata entry
func NewEventStorage(eventID string, taskID a2a.TaskID, eventType string) EventStorage {
	now := time.Now().Unix()
	return EventStorage{
		EventID:    eventID,
		TaskID:     taskID,
		EventType:  eventType,
		StorageKey: fmt.Sprintf("event_%s", eventID),
		Timestamp:  now,
		Processed:  false,
		Metadata:   make(map[string]string),
	}
}

// NewJSONRPCRequest creates a new JSON-RPC request
func NewJSONRPCRequest(method string, params interface{}, id interface{}) JSONRPCRequest {
	return JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      id,
	}
}

// NewJSONRPCResponse creates a new JSON-RPC success response
func NewJSONRPCResponse(result interface{}, id interface{}) JSONRPCResponse {
	return JSONRPCResponse{
		JSONRPC: "2.0",
		Result:  result,
		ID:      id,
	}
}

// NewJSONRPCErrorResponse creates a new JSON-RPC error response
func NewJSONRPCErrorResponse(code int, message string, data interface{}, id interface{}) JSONRPCResponse {
	return JSONRPCResponse{
		JSONRPC: "2.0",
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
		ID: id,
	}
}
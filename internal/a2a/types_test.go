package a2a

import (
	"encoding/json"
	"testing"

	"github.com/a2aproject/a2a-go/a2a"
)

func TestValidateServerlessConfig(t *testing.T) {
	// Test valid config
	validConfig := ServerlessConfig{
		AgentID: "test-agent",
		AgentCard: a2a.AgentCard{
			Name: "Test Agent",
			URL:  "https://example.com/agent",
			Description: "A test agent",
			ProtocolVersion: "1.0",
			Version: "1.0.0",
		},
		CloudConfig: CloudProviderConfig{
			Provider: "aws",
			AWS: &AWSConfig{
				Region:        "us-east-1",
				SQSQueueURL:   "https://sqs.us-east-1.amazonaws.com/123456789/test-queue",
				DynamoDBTable: "test-table",
			},
		},
		LogLevel: "info",
	}
	
	err := ValidateServerlessConfig(validConfig)
	if err != nil {
		t.Errorf("Expected valid config to pass validation, got error: %v", err)
	}
	
	// Test missing agent ID
	invalidConfig := validConfig
	invalidConfig.AgentID = ""
	err = ValidateServerlessConfig(invalidConfig)
	if err == nil {
		t.Error("Expected error for missing agent_id")
	}
	
	// Test missing agent card name
	invalidConfig = validConfig
	invalidConfig.AgentCard.Name = ""
	err = ValidateServerlessConfig(invalidConfig)
	if err == nil {
		t.Error("Expected error for missing agent_card.name")
	}
	
	// Test missing agent card URL
	invalidConfig = validConfig
	invalidConfig.AgentCard.URL = ""
	err = ValidateServerlessConfig(invalidConfig)
	if err == nil {
		t.Error("Expected error for missing agent_card.url")
	}
}

func TestValidateAWSConfig(t *testing.T) {
	// Test valid AWS config
	validConfig := AWSConfig{
		Region:        "us-east-1",
		SQSQueueURL:   "https://sqs.us-east-1.amazonaws.com/123456789/test-queue",
		DynamoDBTable: "test-table",
	}
	
	err := ValidateAWSConfig(validConfig)
	if err != nil {
		t.Errorf("Expected valid AWS config to pass validation, got error: %v", err)
	}
	
	// Test missing region
	invalidConfig := validConfig
	invalidConfig.Region = ""
	err = ValidateAWSConfig(invalidConfig)
	if err == nil {
		t.Error("Expected error for missing region")
	}
	
	// Test missing SQS queue URL
	invalidConfig = validConfig
	invalidConfig.SQSQueueURL = ""
	err = ValidateAWSConfig(invalidConfig)
	if err == nil {
		t.Error("Expected error for missing sqs_queue_url")
	}
	
	// Test missing DynamoDB table
	invalidConfig = validConfig
	invalidConfig.DynamoDBTable = ""
	err = ValidateAWSConfig(invalidConfig)
	if err == nil {
		t.Error("Expected error for missing dynamodb_table")
	}
}

func TestValidateCloudProviderConfig(t *testing.T) {
	// Test valid AWS provider config
	validAWSConfig := CloudProviderConfig{
		Provider: "aws",
		AWS: &AWSConfig{
			Region:        "us-east-1",
			SQSQueueURL:   "https://sqs.us-east-1.amazonaws.com/123456789/test-queue",
			DynamoDBTable: "test-table",
		},
	}
	
	err := ValidateCloudProviderConfig(validAWSConfig)
	if err != nil {
		t.Errorf("Expected valid AWS provider config to pass validation, got error: %v", err)
	}
	
	// Test valid local provider config
	validLocalConfig := CloudProviderConfig{
		Provider: "local",
	}
	
	err = ValidateCloudProviderConfig(validLocalConfig)
	if err != nil {
		t.Errorf("Expected valid local provider config to pass validation, got error: %v", err)
	}
	
	// Test missing provider
	invalidConfig := CloudProviderConfig{}
	err = ValidateCloudProviderConfig(invalidConfig)
	if err == nil {
		t.Error("Expected error for missing provider")
	}
	
	// Test unsupported provider
	invalidConfig = CloudProviderConfig{
		Provider: "unsupported",
	}
	err = ValidateCloudProviderConfig(invalidConfig)
	if err == nil {
		t.Error("Expected error for unsupported provider")
	}
	
	// Test AWS provider without AWS config
	invalidConfig = CloudProviderConfig{
		Provider: "aws",
	}
	err = ValidateCloudProviderConfig(invalidConfig)
	if err == nil {
		t.Error("Expected error for AWS provider without AWS config")
	}
}

func TestValidateJSONRPCRequest(t *testing.T) {
	// Test valid JSON-RPC request
	validRequest := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "test.method",
		Params:  map[string]string{"key": "value"},
		ID:      1,
	}
	
	err := ValidateJSONRPCRequest(validRequest)
	if err != nil {
		t.Errorf("Expected valid JSON-RPC request to pass validation, got error: %v", err)
	}
	
	// Test invalid JSON-RPC version
	invalidRequest := validRequest
	invalidRequest.JSONRPC = "1.0"
	err = ValidateJSONRPCRequest(invalidRequest)
	if err == nil {
		t.Error("Expected error for invalid jsonrpc version")
	}
	
	// Test missing method
	invalidRequest = validRequest
	invalidRequest.Method = ""
	err = ValidateJSONRPCRequest(invalidRequest)
	if err == nil {
		t.Error("Expected error for missing method")
	}
	
	// Test missing ID
	invalidRequest = validRequest
	invalidRequest.ID = nil
	err = ValidateJSONRPCRequest(invalidRequest)
	if err == nil {
		t.Error("Expected error for missing id")
	}
}

func TestJSONSerialization(t *testing.T) {
	// Test ServerlessConfig serialization
	config := ServerlessConfig{
		AgentID: "test-agent",
		AgentCard: a2a.AgentCard{
			Name: "Test Agent",
			URL:  "https://example.com/agent",
			Description: "A test agent",
			ProtocolVersion: "1.0",
			Version: "1.0.0",
		},
		CloudConfig: CloudProviderConfig{
			Provider: "local",
		},
		LogLevel: "info",
	}
	
	jsonBytes, err := ToJSON(config)
	if err != nil {
		t.Errorf("Failed to serialize ServerlessConfig to JSON: %v", err)
	}
	
	var deserializedConfig ServerlessConfig
	err = FromJSON(jsonBytes, &deserializedConfig)
	if err != nil {
		t.Errorf("Failed to deserialize ServerlessConfig from JSON: %v", err)
	}
	
	if deserializedConfig.AgentID != config.AgentID {
		t.Errorf("Expected AgentID %s, got %s", config.AgentID, deserializedConfig.AgentID)
	}
	
	// Test JSONRPCRequest serialization
	request := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "test.method",
		Params:  map[string]string{"key": "value"},
		ID:      1,
	}
	
	jsonBytes, err = ToJSON(request)
	if err != nil {
		t.Errorf("Failed to serialize JSONRPCRequest to JSON: %v", err)
	}
	
	var deserializedRequest JSONRPCRequest
	err = FromJSON(jsonBytes, &deserializedRequest)
	if err != nil {
		t.Errorf("Failed to deserialize JSONRPCRequest from JSON: %v", err)
	}
	
	if deserializedRequest.Method != request.Method {
		t.Errorf("Expected Method %s, got %s", request.Method, deserializedRequest.Method)
	}
}

func TestNewTaskStorage(t *testing.T) {
	taskID := a2a.TaskID("test-task-123")
	contextID := "test-context-456"
	
	storage := NewTaskStorage(taskID, contextID)
	
	if storage.TaskID != taskID {
		t.Errorf("Expected TaskID %s, got %s", taskID, storage.TaskID)
	}
	
	if storage.ContextID != contextID {
		t.Errorf("Expected ContextID %s, got %s", contextID, storage.ContextID)
	}
	
	if storage.StorageKey == "" {
		t.Error("Expected StorageKey to be generated")
	}
	
	if storage.LastModified == 0 {
		t.Error("Expected LastModified to be set")
	}
	
	if storage.Metadata == nil {
		t.Error("Expected Metadata to be initialized")
	}
}

func TestNewEventStorage(t *testing.T) {
	eventID := "test-event-123"
	taskID := a2a.TaskID("test-task-456")
	eventType := "task_created"
	
	storage := NewEventStorage(eventID, taskID, eventType)
	
	if storage.EventID != eventID {
		t.Errorf("Expected EventID %s, got %s", eventID, storage.EventID)
	}
	
	if storage.TaskID != taskID {
		t.Errorf("Expected TaskID %s, got %s", taskID, storage.TaskID)
	}
	
	if storage.EventType != eventType {
		t.Errorf("Expected EventType %s, got %s", eventType, storage.EventType)
	}
	
	if storage.StorageKey == "" {
		t.Error("Expected StorageKey to be generated")
	}
	
	if storage.Timestamp == 0 {
		t.Error("Expected Timestamp to be set")
	}
	
	if storage.Processed {
		t.Error("Expected Processed to be false initially")
	}
	
	if storage.Metadata == nil {
		t.Error("Expected Metadata to be initialized")
	}
}

func TestNewJSONRPCRequest(t *testing.T) {
	method := "test.method"
	params := map[string]string{"key": "value"}
	id := 1
	
	request := NewJSONRPCRequest(method, params, id)
	
	if request.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC '2.0', got %s", request.JSONRPC)
	}
	
	if request.Method != method {
		t.Errorf("Expected Method %s, got %s", method, request.Method)
	}
	
	if request.ID != id {
		t.Errorf("Expected ID %v, got %v", id, request.ID)
	}
}

func TestNewJSONRPCResponse(t *testing.T) {
	result := map[string]string{"result": "success"}
	id := 1
	
	response := NewJSONRPCResponse(result, id)
	
	if response.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC '2.0', got %s", response.JSONRPC)
	}
	
	if response.ID != id {
		t.Errorf("Expected ID %v, got %v", id, response.ID)
	}
	
	if response.Error != nil {
		t.Error("Expected Error to be nil for success response")
	}
}

func TestNewJSONRPCErrorResponse(t *testing.T) {
	code := -32600
	message := "Invalid Request"
	data := map[string]string{"detail": "missing method"}
	id := 1
	
	response := NewJSONRPCErrorResponse(code, message, data, id)
	
	if response.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC '2.0', got %s", response.JSONRPC)
	}
	
	if response.ID != id {
		t.Errorf("Expected ID %v, got %v", id, response.ID)
	}
	
	if response.Error == nil {
		t.Error("Expected Error to be set for error response")
	} else {
		if response.Error.Code != code {
			t.Errorf("Expected Error Code %d, got %d", code, response.Error.Code)
		}
		
		if response.Error.Message != message {
			t.Errorf("Expected Error Message %s, got %s", message, response.Error.Message)
		}
	}
	
	if response.Result != nil {
		t.Error("Expected Result to be nil for error response")
	}
}

func TestTaskStorageJSONSerialization(t *testing.T) {
	storage := NewTaskStorage(a2a.TaskID("test-task"), "test-context")
	
	// Test serialization
	jsonBytes, err := json.Marshal(storage)
	if err != nil {
		t.Errorf("Failed to serialize TaskStorage: %v", err)
	}
	
	// Test deserialization
	var deserializedStorage TaskStorage
	err = json.Unmarshal(jsonBytes, &deserializedStorage)
	if err != nil {
		t.Errorf("Failed to deserialize TaskStorage: %v", err)
	}
	
	// Verify key fields
	if deserializedStorage.TaskID != storage.TaskID {
		t.Errorf("Expected TaskID %s, got %s", storage.TaskID, deserializedStorage.TaskID)
	}
	
	if deserializedStorage.ContextID != storage.ContextID {
		t.Errorf("Expected ContextID %s, got %s", storage.ContextID, deserializedStorage.ContextID)
	}
	
	if deserializedStorage.StorageKey != storage.StorageKey {
		t.Errorf("Expected StorageKey %s, got %s", storage.StorageKey, deserializedStorage.StorageKey)
	}
}
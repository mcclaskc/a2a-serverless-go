package a2a

import (
	"os"
	"strings"
	"testing"

	"github.com/a2aproject/a2a-go/a2a"
)

func TestConfigLoader_LoadServerlessConfig(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid local configuration",
			envVars: map[string]string{
				"A2A_AGENT_ID":          "test-agent-123",
				"A2A_AGENT_NAME":        "Test Agent",
				"A2A_AGENT_URL":         "https://test-agent.example.com",
				"A2A_AGENT_DESCRIPTION": "A test agent for unit testing",
				"A2A_AGENT_VERSION":     "1.0.0",
				"A2A_LOG_LEVEL":         "debug",
				"CLOUD_PROVIDER":        "local",
			},
			expectError: false,
		},
		{
			name: "valid AWS configuration",
			envVars: map[string]string{
				"A2A_AGENT_ID":       "test-agent-aws",
				"A2A_AGENT_NAME":     "AWS Test Agent",
				"A2A_AGENT_URL":      "https://aws-agent.example.com",
				"CLOUD_PROVIDER":     "aws",
				"AWS_REGION":         "us-west-2",
				"AWS_SQS_QUEUE_URL":  "https://sqs.us-west-2.amazonaws.com/123456789/test-queue",
				"AWS_DYNAMODB_TABLE": "test-table",
			},
			expectError: false,
		},
		{
			name: "missing agent ID",
			envVars: map[string]string{
				"A2A_AGENT_NAME": "Test Agent",
				"A2A_AGENT_URL":  "https://test-agent.example.com",
				"CLOUD_PROVIDER": "local",
			},
			expectError: true,
			errorMsg:    "A2A_AGENT_ID environment variable is required",
		},
		{
			name: "missing agent name",
			envVars: map[string]string{
				"A2A_AGENT_ID":   "test-agent-123",
				"A2A_AGENT_URL":  "https://test-agent.example.com",
				"CLOUD_PROVIDER": "local",
			},
			expectError: true,
			errorMsg:    "A2A_AGENT_NAME environment variable is required",
		},
		{
			name: "missing agent URL",
			envVars: map[string]string{
				"A2A_AGENT_ID":   "test-agent-123",
				"A2A_AGENT_NAME": "Test Agent",
				"CLOUD_PROVIDER": "local",
			},
			expectError: true,
			errorMsg:    "A2A_AGENT_URL environment variable is required",
		},
		{
			name: "AWS missing SQS queue URL",
			envVars: map[string]string{
				"A2A_AGENT_ID":       "test-agent-aws",
				"A2A_AGENT_NAME":     "AWS Test Agent",
				"A2A_AGENT_URL":      "https://aws-agent.example.com",
				"CLOUD_PROVIDER":     "aws",
				"AWS_REGION":         "us-west-2",
				"AWS_DYNAMODB_TABLE": "test-table",
			},
			expectError: true,
			errorMsg:    "sqs_queue_url is required",
		},
		{
			name: "AWS missing DynamoDB table",
			envVars: map[string]string{
				"A2A_AGENT_ID":      "test-agent-aws",
				"A2A_AGENT_NAME":    "AWS Test Agent",
				"A2A_AGENT_URL":     "https://aws-agent.example.com",
				"CLOUD_PROVIDER":    "aws",
				"AWS_REGION":        "us-west-2",
				"AWS_SQS_QUEUE_URL": "https://sqs.us-west-2.amazonaws.com/123456789/test-queue",
			},
			expectError: true,
			errorMsg:    "dynamodb_table is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			clearTestEnv()
			
			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			defer clearTestEnv()

			loader := NewConfigLoader()
			config, err := loader.LoadServerlessConfig()

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					// Check if error message contains expected text
					if !containsString(err.Error(), tt.errorMsg) {
						t.Errorf("expected error message to contain '%s', got '%s'", tt.errorMsg, err.Error())
					}
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Validate configuration values
			if config.AgentID != tt.envVars["A2A_AGENT_ID"] {
				t.Errorf("expected AgentID %s, got %s", tt.envVars["A2A_AGENT_ID"], config.AgentID)
			}

			if config.AgentCard.Name != tt.envVars["A2A_AGENT_NAME"] {
				t.Errorf("expected AgentCard.Name %s, got %s", tt.envVars["A2A_AGENT_NAME"], config.AgentCard.Name)
			}

			if config.AgentCard.URL != tt.envVars["A2A_AGENT_URL"] {
				t.Errorf("expected AgentCard.URL %s, got %s", tt.envVars["A2A_AGENT_URL"], config.AgentCard.URL)
			}

			if config.CloudConfig.Provider != tt.envVars["CLOUD_PROVIDER"] {
				t.Errorf("expected CloudConfig.Provider %s, got %s", tt.envVars["CLOUD_PROVIDER"], config.CloudConfig.Provider)
			}
		})
	}
}

func TestConfigLoader_LoadCloudProviderConfig(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		expectError bool
		errorMsg    string
	}{
		{
			name: "local provider",
			envVars: map[string]string{
				"CLOUD_PROVIDER": "local",
			},
			expectError: false,
		},
		{
			name: "AWS provider with valid config",
			envVars: map[string]string{
				"CLOUD_PROVIDER":     "aws",
				"AWS_REGION":         "us-east-1",
				"AWS_SQS_QUEUE_URL":  "https://sqs.us-east-1.amazonaws.com/123456789/test-queue",
				"AWS_DYNAMODB_TABLE": "test-table",
			},
			expectError: false,
		},
		{
			name: "unsupported provider",
			envVars: map[string]string{
				"CLOUD_PROVIDER": "azure",
			},
			expectError: true,
			errorMsg:    "unsupported cloud provider: azure",
		},
		{
			name: "GCP provider (not implemented)",
			envVars: map[string]string{
				"CLOUD_PROVIDER": "gcp",
			},
			expectError: true,
			errorMsg:    "GCP provider not yet implemented",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			clearTestEnv()
			
			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			defer clearTestEnv()

			loader := NewConfigLoader()
			config, err := loader.LoadCloudProviderConfig()

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorMsg != "" && !containsString(err.Error(), tt.errorMsg) {
					t.Errorf("expected error message to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if config.Provider != tt.envVars["CLOUD_PROVIDER"] {
				t.Errorf("expected Provider %s, got %s", tt.envVars["CLOUD_PROVIDER"], config.Provider)
			}
		})
	}
}

func TestConfigLoader_CreateCloudProvider(t *testing.T) {
	tests := []struct {
		name        string
		config      CloudProviderConfig
		expectError bool
		errorMsg    string
		expectType  CloudProvider
	}{
		{
			name: "AWS provider",
			config: CloudProviderConfig{
				Provider: "aws",
				AWS: &AWSConfig{
					Region:        "us-east-1",
					SQSQueueURL:   "https://sqs.us-east-1.amazonaws.com/123456789/test-queue",
					DynamoDBTable: "test-table",
				},
			},
			expectError: false,
			expectType:  CloudProviderAWS,
		},
		{
			name: "local provider",
			config: CloudProviderConfig{
				Provider: "local",
			},
			expectError: false,
			expectType:  CloudProviderLocal,
		},
		{
			name: "AWS provider missing config",
			config: CloudProviderConfig{
				Provider: "aws",
			},
			expectError: true,
			errorMsg:    "AWS configuration is required for AWS provider",
		},
		{
			name: "AWS provider invalid config",
			config: CloudProviderConfig{
				Provider: "aws",
				AWS: &AWSConfig{
					Region: "us-east-1",
					// Missing required fields
				},
			},
			expectError: true,
			errorMsg:    "AWS provider validation failed",
		},
		{
			name: "unsupported provider",
			config: CloudProviderConfig{
				Provider: "azure",
			},
			expectError: true,
			errorMsg:    "unsupported cloud provider: azure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewConfigLoader()
			provider, err := loader.CreateCloudProvider(tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorMsg != "" && !containsString(err.Error(), tt.errorMsg) {
					t.Errorf("expected error message to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if provider.GetProviderType() != tt.expectType {
				t.Errorf("expected provider type %s, got %s", tt.expectType, provider.GetProviderType())
			}
		})
	}
}

func TestAWSProvider(t *testing.T) {
	tests := []struct {
		name        string
		config      AWSConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid AWS config",
			config: AWSConfig{
				Region:        "us-east-1",
				SQSQueueURL:   "https://sqs.us-east-1.amazonaws.com/123456789/test-queue",
				DynamoDBTable: "test-table",
			},
			expectError: false,
		},
		{
			name: "missing region",
			config: AWSConfig{
				SQSQueueURL:   "https://sqs.us-east-1.amazonaws.com/123456789/test-queue",
				DynamoDBTable: "test-table",
			},
			expectError: true,
			errorMsg:    "region is required",
		},
		{
			name: "missing SQS queue URL",
			config: AWSConfig{
				Region:        "us-east-1",
				DynamoDBTable: "test-table",
			},
			expectError: true,
			errorMsg:    "sqs_queue_url is required",
		},
		{
			name: "missing DynamoDB table",
			config: AWSConfig{
				Region:      "us-east-1",
				SQSQueueURL: "https://sqs.us-east-1.amazonaws.com/123456789/test-queue",
			},
			expectError: true,
			errorMsg:    "dynamodb_table is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &AWSProvider{Config: tt.config}
			err := provider.ValidateConfig()

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Test provider methods
			if provider.GetProviderType() != CloudProviderAWS {
				t.Errorf("expected provider type %s, got %s", CloudProviderAWS, provider.GetProviderType())
			}

			storageConfig := provider.GetStorageConfig()
			if storageConfig == nil {
				t.Errorf("expected storage config, got nil")
			}

			eventConfig := provider.GetEventConfig()
			if eventConfig == nil {
				t.Errorf("expected event config, got nil")
			}
		})
	}
}

func TestGCPProvider(t *testing.T) {
	tests := []struct {
		name        string
		provider    GCPProvider
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid GCP config",
			provider: GCPProvider{
				ProjectID:     "test-project",
				FirestoreDB:   "test-db",
				PubSubTopic:   "test-topic",
				Region:        "us-central1",
			},
			expectError: false,
		},
		{
			name: "missing project ID",
			provider: GCPProvider{
				FirestoreDB: "test-db",
				PubSubTopic: "test-topic",
				Region:      "us-central1",
			},
			expectError: true,
			errorMsg:    "gcp project_id is required",
		},
		{
			name: "missing firestore DB",
			provider: GCPProvider{
				ProjectID:   "test-project",
				PubSubTopic: "test-topic",
				Region:      "us-central1",
			},
			expectError: true,
			errorMsg:    "gcp firestore_db is required",
		},
		{
			name: "missing pubsub topic",
			provider: GCPProvider{
				ProjectID:   "test-project",
				FirestoreDB: "test-db",
				Region:      "us-central1",
			},
			expectError: true,
			errorMsg:    "gcp pubsub_topic is required",
		},
		{
			name: "missing region",
			provider: GCPProvider{
				ProjectID:   "test-project",
				FirestoreDB: "test-db",
				PubSubTopic: "test-topic",
			},
			expectError: true,
			errorMsg:    "gcp region is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.provider.ValidateConfig()

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Test provider methods
			if tt.provider.GetProviderType() != CloudProviderGCP {
				t.Errorf("expected provider type %s, got %s", CloudProviderGCP, tt.provider.GetProviderType())
			}

			storageConfig := tt.provider.GetStorageConfig()
			if storageConfig == nil {
				t.Errorf("expected storage config, got nil")
			}

			eventConfig := tt.provider.GetEventConfig()
			if eventConfig == nil {
				t.Errorf("expected event config, got nil")
			}
		})
	}
}

func TestLocalProvider(t *testing.T) {
	provider := &LocalProvider{}
	
	// Test validation (should set defaults)
	err := provider.ValidateConfig()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check defaults were set
	if provider.StoragePath != "./local_storage" {
		t.Errorf("expected default storage path './local_storage', got '%s'", provider.StoragePath)
	}

	if provider.EventPath != "./local_events" {
		t.Errorf("expected default event path './local_events', got '%s'", provider.EventPath)
	}

	// Test provider methods
	if provider.GetProviderType() != CloudProviderLocal {
		t.Errorf("expected provider type %s, got %s", CloudProviderLocal, provider.GetProviderType())
	}

	storageConfig := provider.GetStorageConfig()
	if storageConfig == nil {
		t.Errorf("expected storage config, got nil")
	}

	eventConfig := provider.GetEventConfig()
	if eventConfig == nil {
		t.Errorf("expected event config, got nil")
	}
}

func TestValidateEnvironmentVariables(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		expectError bool
		errorMsg    string
	}{
		{
			name: "all required variables present for local",
			envVars: map[string]string{
				"A2A_AGENT_ID":   "test-agent",
				"A2A_AGENT_NAME": "Test Agent",
				"A2A_AGENT_URL":  "https://test.example.com",
				"CLOUD_PROVIDER": "local",
			},
			expectError: false,
		},
		{
			name: "all required variables present for AWS",
			envVars: map[string]string{
				"A2A_AGENT_ID":       "test-agent",
				"A2A_AGENT_NAME":     "Test Agent",
				"A2A_AGENT_URL":      "https://test.example.com",
				"CLOUD_PROVIDER":     "aws",
				"AWS_SQS_QUEUE_URL":  "https://sqs.us-east-1.amazonaws.com/123456789/test-queue",
				"AWS_DYNAMODB_TABLE": "test-table",
			},
			expectError: false,
		},
		{
			name: "missing required A2A variables",
			envVars: map[string]string{
				"CLOUD_PROVIDER": "local",
			},
			expectError: true,
			errorMsg:    "missing required environment variables: A2A_AGENT_ID, A2A_AGENT_NAME, A2A_AGENT_URL",
		},
		{
			name: "missing AWS-specific variables",
			envVars: map[string]string{
				"A2A_AGENT_ID":   "test-agent",
				"A2A_AGENT_NAME": "Test Agent",
				"A2A_AGENT_URL":  "https://test.example.com",
				"CLOUD_PROVIDER": "aws",
			},
			expectError: true,
			errorMsg:    "missing required environment variables for aws provider: AWS_SQS_QUEUE_URL, AWS_DYNAMODB_TABLE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			clearTestEnv()
			
			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			defer clearTestEnv()

			err := ValidateEnvironmentVariables()

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestLoadAgentCard(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		expectError bool
		errorMsg    string
		expected    a2a.AgentCard
	}{
		{
			name: "complete agent card",
			envVars: map[string]string{
				"A2A_AGENT_NAME":                "Test Agent",
				"A2A_AGENT_URL":                 "https://test.example.com",
				"A2A_AGENT_DESCRIPTION":         "A test agent",
				"A2A_AGENT_VERSION":             "2.0.0",
				"A2A_AGENT_PUSH_NOTIFICATIONS": "true",
				"A2A_AGENT_STATE_HISTORY":       "true",
				"A2A_AGENT_STREAMING":           "false",
			},
			expectError: false,
			expected: a2a.AgentCard{
				Name:        "Test Agent",
				URL:         "https://test.example.com",
				Description: "A test agent",
				Version:     "2.0.0",
				Capabilities: a2a.AgentCapabilities{
					PushNotifications:       boolPtr(true),
					StateTransitionHistory:  boolPtr(true),
					Streaming:               boolPtr(false),
				},
			},
		},
		{
			name: "minimal agent card with defaults",
			envVars: map[string]string{
				"A2A_AGENT_NAME": "Minimal Agent",
				"A2A_AGENT_URL":  "https://minimal.example.com",
			},
			expectError: false,
			expected: a2a.AgentCard{
				Name:        "Minimal Agent",
				URL:         "https://minimal.example.com",
				Description: "",
				Version:     "1.0.0",
				Capabilities: a2a.AgentCapabilities{},
			},
		},
		{
			name: "missing agent name",
			envVars: map[string]string{
				"A2A_AGENT_URL": "https://test.example.com",
			},
			expectError: true,
			errorMsg:    "A2A_AGENT_NAME environment variable is required",
		},
		{
			name: "missing agent URL",
			envVars: map[string]string{
				"A2A_AGENT_NAME": "Test Agent",
			},
			expectError: true,
			errorMsg:    "A2A_AGENT_URL environment variable is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			clearTestEnv()
			
			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			defer clearTestEnv()

			loader := NewConfigLoader()
			agentCard, err := loader.loadAgentCard()

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Compare agent card fields
			if agentCard.Name != tt.expected.Name {
				t.Errorf("expected Name '%s', got '%s'", tt.expected.Name, agentCard.Name)
			}
			if agentCard.URL != tt.expected.URL {
				t.Errorf("expected URL '%s', got '%s'", tt.expected.URL, agentCard.URL)
			}
			if agentCard.Description != tt.expected.Description {
				t.Errorf("expected Description '%s', got '%s'", tt.expected.Description, agentCard.Description)
			}
			if agentCard.Version != tt.expected.Version {
				t.Errorf("expected Version '%s', got '%s'", tt.expected.Version, agentCard.Version)
			}

			// Compare capabilities
			if !compareCapabilities(agentCard.Capabilities, tt.expected.Capabilities) {
				t.Errorf("capabilities mismatch: expected %+v, got %+v", tt.expected.Capabilities, agentCard.Capabilities)
			}
		})
	}
}

// Helper functions for tests

func clearTestEnv() {
	envVars := []string{
		"A2A_AGENT_ID", "A2A_AGENT_NAME", "A2A_AGENT_URL", "A2A_AGENT_DESCRIPTION",
		"A2A_AGENT_VERSION", "A2A_AGENT_PUSH_NOTIFICATIONS", "A2A_AGENT_STATE_HISTORY", 
		"A2A_AGENT_STREAMING", "A2A_LOG_LEVEL",
		"CLOUD_PROVIDER", "AWS_REGION", "AWS_SQS_QUEUE_URL", "AWS_DYNAMODB_TABLE",
		"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY",
		"GCP_PROJECT_ID", "GCP_FIRESTORE_DB", "GCP_PUBSUB_TOPIC",
		"LOCAL_STORAGE_PATH", "LOCAL_EVENT_PATH",
	}
	
	for _, env := range envVars {
		os.Unsetenv(env)
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && 
			(s[:len(substr)] == substr || 
			 s[len(s)-len(substr):] == substr ||
			 strings.Contains(s, substr))))
}

func boolPtr(b bool) *bool {
	return &b
}

func compareCapabilities(a, b a2a.AgentCapabilities) bool {
	// Compare PushNotifications
	if (a.PushNotifications == nil) != (b.PushNotifications == nil) {
		return false
	}
	if a.PushNotifications != nil && b.PushNotifications != nil && *a.PushNotifications != *b.PushNotifications {
		return false
	}
	
	// Compare StateTransitionHistory
	if (a.StateTransitionHistory == nil) != (b.StateTransitionHistory == nil) {
		return false
	}
	if a.StateTransitionHistory != nil && b.StateTransitionHistory != nil && *a.StateTransitionHistory != *b.StateTransitionHistory {
		return false
	}
	
	// Compare Streaming
	if (a.Streaming == nil) != (b.Streaming == nil) {
		return false
	}
	if a.Streaming != nil && b.Streaming != nil && *a.Streaming != *b.Streaming {
		return false
	}
	
	// Compare Extensions (length should be same for empty slices)
	if len(a.Extensions) != len(b.Extensions) {
		return false
	}
	
	return true
}
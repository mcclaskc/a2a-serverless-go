package a2a

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/a2aproject/a2a-go/a2a"
)

// CloudProvider represents different cloud provider types
type CloudProvider string

const (
	CloudProviderAWS   CloudProvider = "aws"
	CloudProviderGCP   CloudProvider = "gcp"
	CloudProviderLocal CloudProvider = "local"
)

// CloudProviderInterface defines the interface for cloud provider operations
type CloudProviderInterface interface {
	// GetProviderType returns the provider type
	GetProviderType() CloudProvider
	
	// ValidateConfig validates the provider-specific configuration
	ValidateConfig() error
	
	// GetStorageConfig returns storage configuration for the provider
	GetStorageConfig() interface{}
	
	// GetEventConfig returns event queue configuration for the provider
	GetEventConfig() interface{}
}

// AWSProvider implements CloudProviderInterface for AWS
type AWSProvider struct {
	Config AWSConfig
}

// GetProviderType returns AWS provider type
func (p *AWSProvider) GetProviderType() CloudProvider {
	return CloudProviderAWS
}

// ValidateConfig validates AWS configuration
func (p *AWSProvider) ValidateConfig() error {
	return ValidateAWSConfig(p.Config)
}

// GetStorageConfig returns AWS DynamoDB configuration
func (p *AWSProvider) GetStorageConfig() interface{} {
	return map[string]string{
		"table_name": p.Config.DynamoDBTable,
		"region":     p.Config.Region,
	}
}

// GetEventConfig returns AWS SQS configuration
func (p *AWSProvider) GetEventConfig() interface{} {
	return map[string]string{
		"queue_url": p.Config.SQSQueueURL,
		"region":    p.Config.Region,
	}
}

// GCPProvider implements CloudProviderInterface for GCP
type GCPProvider struct {
	ProjectID     string
	FirestoreDB   string
	PubSubTopic   string
	Region        string
	CredentialsPath string
}

// GetProviderType returns GCP provider type
func (p *GCPProvider) GetProviderType() CloudProvider {
	return CloudProviderGCP
}

// ValidateConfig validates GCP configuration
func (p *GCPProvider) ValidateConfig() error {
	if p.ProjectID == "" {
		return fmt.Errorf("gcp project_id is required")
	}
	if p.FirestoreDB == "" {
		return fmt.Errorf("gcp firestore_db is required")
	}
	if p.PubSubTopic == "" {
		return fmt.Errorf("gcp pubsub_topic is required")
	}
	if p.Region == "" {
		return fmt.Errorf("gcp region is required")
	}
	return nil
}

// GetStorageConfig returns GCP Firestore configuration
func (p *GCPProvider) GetStorageConfig() interface{} {
	return map[string]string{
		"project_id":       p.ProjectID,
		"firestore_db":     p.FirestoreDB,
		"region":           p.Region,
		"credentials_path": p.CredentialsPath,
	}
}

// GetEventConfig returns GCP Pub/Sub configuration
func (p *GCPProvider) GetEventConfig() interface{} {
	return map[string]string{
		"project_id":       p.ProjectID,
		"pubsub_topic":     p.PubSubTopic,
		"region":           p.Region,
		"credentials_path": p.CredentialsPath,
	}
}

// LocalProvider implements CloudProviderInterface for local development
type LocalProvider struct {
	StoragePath string
	EventPath   string
}

// GetProviderType returns local provider type
func (p *LocalProvider) GetProviderType() CloudProvider {
	return CloudProviderLocal
}

// ValidateConfig validates local configuration
func (p *LocalProvider) ValidateConfig() error {
	// Local provider has minimal validation requirements
	if p.StoragePath == "" {
		p.StoragePath = "./local_storage"
	}
	if p.EventPath == "" {
		p.EventPath = "./local_events"
	}
	return nil
}

// GetStorageConfig returns local storage configuration
func (p *LocalProvider) GetStorageConfig() interface{} {
	return map[string]string{
		"storage_path": p.StoragePath,
	}
}

// GetEventConfig returns local event configuration
func (p *LocalProvider) GetEventConfig() interface{} {
	return map[string]string{
		"event_path": p.EventPath,
	}
}

// ConfigLoader handles loading configuration from environment variables
type ConfigLoader struct{}

// NewConfigLoader creates a new configuration loader
func NewConfigLoader() *ConfigLoader {
	return &ConfigLoader{}
}

// LoadServerlessConfig loads complete serverless configuration from environment
func (cl *ConfigLoader) LoadServerlessConfig() (ServerlessConfig, error) {
	// Load basic A2A configuration
	agentID := getEnvOrDefault("A2A_AGENT_ID", "")
	if agentID == "" {
		return ServerlessConfig{}, fmt.Errorf("A2A_AGENT_ID environment variable is required")
	}

	// Load agent card configuration
	agentCard, err := cl.loadAgentCard()
	if err != nil {
		return ServerlessConfig{}, fmt.Errorf("failed to load agent card: %w", err)
	}

	// Load cloud provider configuration
	cloudConfig, err := cl.LoadCloudProviderConfig()
	if err != nil {
		return ServerlessConfig{}, fmt.Errorf("failed to load cloud provider config: %w", err)
	}

	// Load logging configuration
	logLevel := getEnvOrDefault("A2A_LOG_LEVEL", "info")

	config := ServerlessConfig{
		AgentID:     agentID,
		AgentCard:   agentCard,
		CloudConfig: cloudConfig,
		LogLevel:    logLevel,
	}

	// Validate the complete configuration
	if err := ValidateServerlessConfig(config); err != nil {
		return ServerlessConfig{}, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// LoadCloudProviderConfig loads cloud provider configuration from environment
func (cl *ConfigLoader) LoadCloudProviderConfig() (CloudProviderConfig, error) {
	provider := getEnvOrDefault("CLOUD_PROVIDER", "local")
	
	switch CloudProvider(provider) {
	case CloudProviderAWS:
		awsConfig, err := cl.loadAWSConfig()
		if err != nil {
			return CloudProviderConfig{}, fmt.Errorf("failed to load AWS config: %w", err)
		}
		return CloudProviderConfig{
			Provider: provider,
			AWS:      &awsConfig,
		}, nil
		
	case CloudProviderGCP:
		// GCP configuration will be implemented in future tasks
		return CloudProviderConfig{}, fmt.Errorf("GCP provider not yet implemented")
		
	case CloudProviderLocal:
		return CloudProviderConfig{
			Provider: provider,
		}, nil
		
	default:
		return CloudProviderConfig{}, fmt.Errorf("unsupported cloud provider: %s", provider)
	}
}

// CreateCloudProvider creates a cloud provider instance based on configuration
func (cl *ConfigLoader) CreateCloudProvider(config CloudProviderConfig) (CloudProviderInterface, error) {
	switch CloudProvider(config.Provider) {
	case CloudProviderAWS:
		if config.AWS == nil {
			return nil, fmt.Errorf("AWS configuration is required for AWS provider")
		}
		provider := &AWSProvider{Config: *config.AWS}
		if err := provider.ValidateConfig(); err != nil {
			return nil, fmt.Errorf("AWS provider validation failed: %w", err)
		}
		return provider, nil
		
	case CloudProviderGCP:
		// GCP provider will be implemented in future tasks
		return nil, fmt.Errorf("GCP provider not yet implemented")
		
	case CloudProviderLocal:
		provider := &LocalProvider{
			StoragePath: getEnvOrDefault("LOCAL_STORAGE_PATH", "./local_storage"),
			EventPath:   getEnvOrDefault("LOCAL_EVENT_PATH", "./local_events"),
		}
		if err := provider.ValidateConfig(); err != nil {
			return nil, fmt.Errorf("local provider validation failed: %w", err)
		}
		return provider, nil
		
	default:
		return nil, fmt.Errorf("unsupported cloud provider: %s", config.Provider)
	}
}

// loadAgentCard loads agent card configuration from environment variables
func (cl *ConfigLoader) loadAgentCard() (a2a.AgentCard, error) {
	name := getEnvOrDefault("A2A_AGENT_NAME", "")
	if name == "" {
		return a2a.AgentCard{}, fmt.Errorf("A2A_AGENT_NAME environment variable is required")
	}

	url := getEnvOrDefault("A2A_AGENT_URL", "")
	if url == "" {
		return a2a.AgentCard{}, fmt.Errorf("A2A_AGENT_URL environment variable is required")
	}

	description := getEnvOrDefault("A2A_AGENT_DESCRIPTION", "")
	version := getEnvOrDefault("A2A_AGENT_VERSION", "1.0.0")
	
	// Parse capabilities configuration
	capabilities := a2a.AgentCapabilities{}
	
	// Parse boolean capabilities from environment variables
	// Only set the pointer if the environment variable is explicitly set
	if os.Getenv("A2A_AGENT_PUSH_NOTIFICATIONS") != "" {
		pushNotifications := getEnvOrDefaultBool("A2A_AGENT_PUSH_NOTIFICATIONS", false)
		capabilities.PushNotifications = &pushNotifications
	}
	
	if os.Getenv("A2A_AGENT_STATE_HISTORY") != "" {
		stateHistory := getEnvOrDefaultBool("A2A_AGENT_STATE_HISTORY", false)
		capabilities.StateTransitionHistory = &stateHistory
	}
	
	if os.Getenv("A2A_AGENT_STREAMING") != "" {
		streaming := getEnvOrDefaultBool("A2A_AGENT_STREAMING", false)
		capabilities.Streaming = &streaming
	}

	return a2a.AgentCard{
		Name:         name,
		URL:          url,
		Description:  description,
		Version:      version,
		Capabilities: capabilities,
	}, nil
}

// loadAWSConfig loads AWS configuration from environment variables
func (cl *ConfigLoader) loadAWSConfig() (AWSConfig, error) {
	region := getEnvOrDefault("AWS_REGION", "us-east-1")
	sqsQueueURL := getEnvOrDefault("AWS_SQS_QUEUE_URL", "")
	dynamoDBTable := getEnvOrDefault("AWS_DYNAMODB_TABLE", "")
	
	// Optional credentials (can use IAM roles instead)
	accessKeyID := getEnvOrDefault("AWS_ACCESS_KEY_ID", "")
	secretAccessKey := getEnvOrDefault("AWS_SECRET_ACCESS_KEY", "")

	config := AWSConfig{
		Region:          region,
		SQSQueueURL:     sqsQueueURL,
		DynamoDBTable:   dynamoDBTable,
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
	}

	return config, nil
}

// getEnvOrDefault gets environment variable value or returns default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvOrDefaultInt gets environment variable as integer or returns default
func getEnvOrDefaultInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvOrDefaultBool gets environment variable as boolean or returns default
func getEnvOrDefaultBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// ValidateEnvironmentVariables validates that required environment variables are set
func ValidateEnvironmentVariables() error {
	required := []string{
		"A2A_AGENT_ID",
		"A2A_AGENT_NAME", 
		"A2A_AGENT_URL",
	}

	var missing []string
	for _, env := range required {
		if os.Getenv(env) == "" {
			missing = append(missing, env)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	// Validate provider-specific requirements
	provider := getEnvOrDefault("CLOUD_PROVIDER", "local")
	switch CloudProvider(provider) {
	case CloudProviderAWS:
		awsRequired := []string{"AWS_SQS_QUEUE_URL", "AWS_DYNAMODB_TABLE"}
		for _, env := range awsRequired {
			if os.Getenv(env) == "" {
				missing = append(missing, env)
			}
		}
	case CloudProviderGCP:
		gcpRequired := []string{"GCP_PROJECT_ID", "GCP_FIRESTORE_DB", "GCP_PUBSUB_TOPIC"}
		for _, env := range gcpRequired {
			if os.Getenv(env) == "" {
				missing = append(missing, env)
			}
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables for %s provider: %s", provider, strings.Join(missing, ", "))
	}

	return nil
}
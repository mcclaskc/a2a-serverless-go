package a2a

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/a2aproject/a2a-go/a2a"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// AWSTaskStore implements TaskStore using DynamoDB
type AWSTaskStore struct {
	client    *dynamodb.Client
	tableName string
}

// NewAWSTaskStore creates a new AWS DynamoDB-based task store
func NewAWSTaskStore(client *dynamodb.Client, tableName string) *AWSTaskStore {
	return &AWSTaskStore{
		client:    client,
		tableName: tableName,
	}
}

// GetTask retrieves a task from DynamoDB
func (s *AWSTaskStore) GetTask(ctx context.Context, taskID a2a.TaskID) (a2a.Task, error) {
	result, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"task_id": &types.AttributeValueMemberS{Value: string(taskID)},
		},
	})
	if err != nil {
		return a2a.Task{}, fmt.Errorf("failed to get task from DynamoDB: %w", err)
	}

	if result.Item == nil {
		return a2a.Task{}, fmt.Errorf("task %s not found", taskID)
	}

	// Extract task data from DynamoDB item
	taskDataAttr, ok := result.Item["task_data"]
	if !ok {
		return a2a.Task{}, fmt.Errorf("task_data not found in DynamoDB item")
	}

	taskDataStr, ok := taskDataAttr.(*types.AttributeValueMemberS)
	if !ok {
		return a2a.Task{}, fmt.Errorf("task_data is not a string")
	}

	var task a2a.Task
	err = json.Unmarshal([]byte(taskDataStr.Value), &task)
	if err != nil {
		return a2a.Task{}, fmt.Errorf("failed to unmarshal task data: %w", err)
	}

	return task, nil
}

// SaveTask saves a task to DynamoDB
func (s *AWSTaskStore) SaveTask(ctx context.Context, task a2a.Task) error {
	taskData, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	_, err = s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.tableName),
		Item: map[string]types.AttributeValue{
			"task_id": &types.AttributeValueMemberS{Value: string(task.ID)},
			"context_id": &types.AttributeValueMemberS{Value: task.ContextID},
			"task_data": &types.AttributeValueMemberS{Value: string(taskData)},
			"status": &types.AttributeValueMemberS{Value: string(task.Status.State)},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to save task to DynamoDB: %w", err)
	}

	return nil
}

// DeleteTask deletes a task from DynamoDB
func (s *AWSTaskStore) DeleteTask(ctx context.Context, taskID a2a.TaskID) error {
	_, err := s.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"task_id": &types.AttributeValueMemberS{Value: string(taskID)},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete task from DynamoDB: %w", err)
	}

	return nil
}

// ListTasks lists tasks by context ID from DynamoDB
func (s *AWSTaskStore) ListTasks(ctx context.Context, contextID string) ([]a2a.Task, error) {
	result, err := s.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(s.tableName),
		IndexName:              aws.String("context_id-index"), // Assumes GSI exists
		KeyConditionExpression: aws.String("context_id = :context_id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":context_id": &types.AttributeValueMemberS{Value: contextID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks from DynamoDB: %w", err)
	}

	var tasks []a2a.Task
	for _, item := range result.Items {
		taskDataAttr, ok := item["task_data"]
		if !ok {
			continue
		}

		taskDataStr, ok := taskDataAttr.(*types.AttributeValueMemberS)
		if !ok {
			continue
		}

		var task a2a.Task
		err = json.Unmarshal([]byte(taskDataStr.Value), &task)
		if err != nil {
			// Log error but continue with other tasks
			continue
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

// AWSEventStore implements EventStore using DynamoDB
type AWSEventStore struct {
	client    *dynamodb.Client
	tableName string
}

// NewAWSEventStore creates a new AWS DynamoDB-based event store
func NewAWSEventStore(client *dynamodb.Client, tableName string) *AWSEventStore {
	return &AWSEventStore{
		client:    client,
		tableName: tableName,
	}
}

// SaveEvent saves an event to DynamoDB
func (s *AWSEventStore) SaveEvent(ctx context.Context, event a2a.Event) error {
	eventData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Generate event ID based on event type
	var eventID string
	var taskID a2a.TaskID

	switch e := event.(type) {
	case a2a.TaskStatusUpdateEvent:
		eventID = fmt.Sprintf("status_%s_%d", e.TaskID, e.Status.Timestamp.UnixNano())
		taskID = e.TaskID
	case a2a.TaskArtifactUpdateEvent:
		eventID = fmt.Sprintf("artifact_%s_%s", e.TaskID, e.Artifact.ArtifactID)
		taskID = e.TaskID
	case a2a.Message:
		eventID = e.MessageID
		if e.TaskID != nil {
			taskID = *e.TaskID
		}
	default:
		eventID = fmt.Sprintf("event_%d", time.Now().UnixNano())
	}

	_, err = s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.tableName),
		Item: map[string]types.AttributeValue{
			"event_id": &types.AttributeValueMemberS{Value: eventID},
			"task_id": &types.AttributeValueMemberS{Value: string(taskID)},
			"event_data": &types.AttributeValueMemberS{Value: string(eventData)},
			"processed": &types.AttributeValueMemberBOOL{Value: false},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to save event to DynamoDB: %w", err)
	}

	return nil
}

// GetEvents retrieves events for a task from DynamoDB
func (s *AWSEventStore) GetEvents(ctx context.Context, taskID a2a.TaskID) ([]a2a.Event, error) {
	result, err := s.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(s.tableName),
		IndexName:              aws.String("task_id-index"), // Assumes GSI exists
		KeyConditionExpression: aws.String("task_id = :task_id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":task_id": &types.AttributeValueMemberS{Value: string(taskID)},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query events from DynamoDB: %w", err)
	}

	var events []a2a.Event
	for _, item := range result.Items {
		eventDataAttr, ok := item["event_data"]
		if !ok {
			continue
		}

		eventDataStr, ok := eventDataAttr.(*types.AttributeValueMemberS)
		if !ok {
			continue
		}

		// Parse the event data to determine type
		var eventData map[string]interface{}
		err = json.Unmarshal([]byte(eventDataStr.Value), &eventData)
		if err != nil {
			continue
		}

		// Convert to appropriate event type based on "kind" field
		kind, ok := eventData["kind"].(string)
		if !ok {
			continue
		}

		var event a2a.Event
		switch kind {
		case "status-update":
			var statusEvent a2a.TaskStatusUpdateEvent
			err = json.Unmarshal([]byte(eventDataStr.Value), &statusEvent)
			if err == nil {
				event = statusEvent
			}
		case "artifact-update":
			var artifactEvent a2a.TaskArtifactUpdateEvent
			err = json.Unmarshal([]byte(eventDataStr.Value), &artifactEvent)
			if err == nil {
				event = artifactEvent
			}
		case "message":
			var message a2a.Message
			err = json.Unmarshal([]byte(eventDataStr.Value), &message)
			if err == nil {
				event = message
			}
		default:
			// Skip unknown event types
			continue
		}

		if event != nil {
			events = append(events, event)
		}
	}

	return events, nil
}

// MarkEventProcessed marks an event as processed in DynamoDB
func (s *AWSEventStore) MarkEventProcessed(ctx context.Context, eventID string) error {
	_, err := s.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"event_id": &types.AttributeValueMemberS{Value: eventID},
		},
		UpdateExpression: aws.String("SET processed = :processed"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":processed": &types.AttributeValueMemberBOOL{Value: true},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to mark event as processed: %w", err)
	}

	return nil
}

// AWSSQSPushNotifier implements PushNotifier using SQS
type AWSSQSPushNotifier struct {
	client   *sqs.Client
	queueURL string
}

// NewAWSSQSPushNotifier creates a new AWS SQS-based push notifier
func NewAWSSQSPushNotifier(client *sqs.Client, queueURL string) *AWSSQSPushNotifier {
	return &AWSSQSPushNotifier{
		client:   client,
		queueURL: queueURL,
	}
}

// SendNotification sends a push notification via SQS
func (n *AWSSQSPushNotifier) SendNotification(ctx context.Context, config a2a.PushConfig, event a2a.Event) error {
	notification := map[string]interface{}{
		"push_config": config,
		"event":       event,
	}

	notificationData, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	_, err = n.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(n.queueURL),
		MessageBody: aws.String(string(notificationData)),
	})
	if err != nil {
		return fmt.Errorf("failed to send notification to SQS: %w", err)
	}

	return nil
}
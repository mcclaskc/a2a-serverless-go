package a2a

import (
	"context"
	"fmt"
	"iter"
	"time"

	"github.com/a2aproject/a2a-go/a2a"
	"github.com/a2aproject/a2a-go/a2asrv"
)

// ServerlessA2AHandler implements the A2A RequestHandler interface for serverless environments
type ServerlessA2AHandler struct {
	config       ServerlessConfig
	taskStore    TaskStore
	eventStore   EventStore
	pushNotifier PushNotifier
}

// TaskStore defines the interface for task persistence in serverless environments
type TaskStore interface {
	GetTask(ctx context.Context, taskID a2a.TaskID) (a2a.Task, error)
	SaveTask(ctx context.Context, task a2a.Task) error
	DeleteTask(ctx context.Context, taskID a2a.TaskID) error
	ListTasks(ctx context.Context, contextID string) ([]a2a.Task, error)
}

// EventStore defines the interface for event persistence in serverless environments
type EventStore interface {
	SaveEvent(ctx context.Context, event a2a.Event) error
	GetEvents(ctx context.Context, taskID a2a.TaskID) ([]a2a.Event, error)
	MarkEventProcessed(ctx context.Context, eventID string) error
}

// PushNotifier defines the interface for sending push notifications
type PushNotifier interface {
	SendNotification(ctx context.Context, config a2a.PushConfig, event a2a.Event) error
}

// NewServerlessA2AHandler creates a new serverless A2A handler
func NewServerlessA2AHandler(config ServerlessConfig, taskStore TaskStore, eventStore EventStore, pushNotifier PushNotifier) *ServerlessA2AHandler {
	return &ServerlessA2AHandler{
		config:       config,
		taskStore:    taskStore,
		eventStore:   eventStore,
		pushNotifier: pushNotifier,
	}
}

// Verify that ServerlessA2AHandler implements the RequestHandler interface
var _ a2asrv.RequestHandler = (*ServerlessA2AHandler)(nil)

// OnGetTask handles the 'tasks/get' protocol method
func (h *ServerlessA2AHandler) OnGetTask(ctx context.Context, query a2a.TaskQueryParams) (a2a.Task, error) {
	task, err := h.taskStore.GetTask(ctx, query.ID)
	if err != nil {
		return a2a.Task{}, fmt.Errorf("failed to get task %s: %w", query.ID, err)
	}

	// Limit history if requested
	if query.HistoryLength != nil && *query.HistoryLength > 0 {
		historyLen := *query.HistoryLength
		if len(task.History) > historyLen {
			task.History = task.History[len(task.History)-historyLen:]
		}
	}

	return task, nil
}

// OnCancelTask handles the 'tasks/cancel' protocol method
func (h *ServerlessA2AHandler) OnCancelTask(ctx context.Context, id a2a.TaskIDParams) (a2a.Task, error) {
	task, err := h.taskStore.GetTask(ctx, id.ID)
	if err != nil {
		return a2a.Task{}, fmt.Errorf("failed to get task %s: %w", id.ID, err)
	}

	// Update task status to canceled
	now := time.Now()
	task.Status = a2a.TaskStatus{
		State:     a2a.TaskStateCanceled,
		Timestamp: &now,
	}

	err = h.taskStore.SaveTask(ctx, task)
	if err != nil {
		return a2a.Task{}, fmt.Errorf("failed to save canceled task %s: %w", id.ID, err)
	}

	// Create and store status update event
	statusEvent := a2a.TaskStatusUpdateEvent{
		Kind:      "status-update",
		TaskID:    task.ID,
		ContextID: task.ContextID,
		Status:    task.Status,
		Final:     true,
	}

	err = h.eventStore.SaveEvent(ctx, statusEvent)
	if err != nil {
		// Log error but don't fail the request
		// In a real implementation, you'd use proper logging
		fmt.Printf("Warning: failed to save status event for task %s: %v\n", id.ID, err)
	}

	return task, nil
}

// OnSendMessage handles the 'message/send' protocol method (non-streaming)
func (h *ServerlessA2AHandler) OnSendMessage(ctx context.Context, message a2a.MessageSendParams) (a2a.SendMessageResult, error) {
	// This is a simplified implementation - in a real serverless environment,
	// you would likely queue the message for processing by another function
	
	var task a2a.Task
	var err error

	if message.Message.TaskID != nil {
		// Continue existing task
		task, err = h.taskStore.GetTask(ctx, *message.Message.TaskID)
		if err != nil {
			return nil, fmt.Errorf("failed to get existing task %s: %w", *message.Message.TaskID, err)
		}
	} else {
		// Create new task
		now := time.Now()
		task = a2a.Task{
			ID:        a2a.TaskID(fmt.Sprintf("task_%d", now.UnixNano())),
			ContextID: generateContextID(),
			Kind:      "task",
			History:   []a2a.Message{},
			Status: a2a.TaskStatus{
				State:     a2a.TaskStateSubmitted,
				Timestamp: &now,
			},
			Metadata: make(map[string]any),
		}
	}

	// Add message to task history
	task.History = append(task.History, message.Message)

	// Update task status to working
	now := time.Now()
	task.Status = a2a.TaskStatus{
		State:     a2a.TaskStateWorking,
		Timestamp: &now,
	}

	// Save updated task
	err = h.taskStore.SaveTask(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("failed to save task: %w", err)
	}

	// In a real implementation, you would process the message here
	// For now, we'll just return the task
	return task, nil
}

// OnResubscribeToTask handles the `tasks/resubscribe` protocol method
func (h *ServerlessA2AHandler) OnResubscribeToTask(ctx context.Context, id a2a.TaskIDParams) iter.Seq2[a2a.Event, error] {
	return func(yield func(a2a.Event, error) bool) {
		events, err := h.eventStore.GetEvents(ctx, id.ID)
		if err != nil {
			yield(nil, fmt.Errorf("failed to get events for task %s: %w", id.ID, err))
			return
		}

		for _, event := range events {
			if !yield(event, nil) {
				return
			}
		}
	}
}

// OnSendMessageStream handles the 'message/stream' protocol method (streaming)
func (h *ServerlessA2AHandler) OnSendMessageStream(ctx context.Context, message a2a.MessageSendParams) iter.Seq2[a2a.Event, error] {
	return func(yield func(a2a.Event, error) bool) {
		// First, handle the message like in OnSendMessage
		result, err := h.OnSendMessage(ctx, message)
		if err != nil {
			yield(nil, err)
			return
		}

		// Convert result to appropriate event
		if task, ok := result.(a2a.Task); ok {
			// Send status update event
			statusEvent := a2a.TaskStatusUpdateEvent{
				Kind:      "status-update",
				TaskID:    task.ID,
				ContextID: task.ContextID,
				Status:    task.Status,
				Final:     false,
			}

			if !yield(statusEvent, nil) {
				return
			}

			// In a real implementation, you would continue yielding events
			// as the task progresses
		}
	}
}

// OnGetTaskPushConfig handles the `tasks/pushNotificationConfig/get` protocol method
func (h *ServerlessA2AHandler) OnGetTaskPushConfig(ctx context.Context, params a2a.GetTaskPushConfigParams) (a2a.TaskPushConfig, error) {
	// This would typically be stored in a database
	// For now, return an empty config
	return a2a.TaskPushConfig{
		TaskID: params.TaskID,
		Config: a2a.PushConfig{},
	}, nil
}

// OnListTaskPushConfig handles the `tasks/pushNotificationConfig/list` protocol method
func (h *ServerlessA2AHandler) OnListTaskPushConfig(ctx context.Context, params a2a.ListTaskPushConfigParams) ([]a2a.TaskPushConfig, error) {
	// This would typically be stored in a database
	// For now, return an empty list
	return []a2a.TaskPushConfig{}, nil
}

// OnSetTaskPushConfig handles the `tasks/pushNotificationConfig/set` protocol method
func (h *ServerlessA2AHandler) OnSetTaskPushConfig(ctx context.Context, params a2a.TaskPushConfig) (a2a.TaskPushConfig, error) {
	// This would typically be stored in a database
	// For now, just return the input
	return params, nil
}

// OnDeleteTaskPushConfig handles the `tasks/pushNotificationConfig/delete` protocol method
func (h *ServerlessA2AHandler) OnDeleteTaskPushConfig(ctx context.Context, params a2a.DeleteTaskPushConfigParams) error {
	// This would typically delete from a database
	// For now, just return success
	return nil
}

// generateContextID generates a unique context ID
func generateContextID() string {
	return fmt.Sprintf("ctx_%d", time.Now().UnixNano())
}
package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/a2aproject/a2a-go/a2a"
	a2aTypes "github.com/a2aproject/a2a-serverless/internal/a2a"
)

// Request represents an incoming HTTP request
type Request struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

// Response represents an HTTP response
type Response struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

// Handler contains the A2A serverless handler
type Handler struct {
	a2aHandler *a2aTypes.ServerlessA2AHandler
	agentCard  a2a.AgentCard
}

// NewHandler creates a new handler instance with A2A support
func NewHandler(a2aHandler *a2aTypes.ServerlessA2AHandler, agentCard a2a.AgentCard) *Handler {
	return &Handler{
		a2aHandler: a2aHandler,
		agentCard:  agentCard,
	}
}

// HandleRequest processes incoming requests - routes to A2A or returns agent card
func (h *Handler) HandleRequest(req Request) Response {
	ctx := context.Background()

	// Handle CORS preflight requests
	if req.Method == "OPTIONS" {
		return h.handleCORS()
	}

	// Handle agent card requests
	if req.Method == "GET" && (req.URL == "/" || req.URL == "/agent-card") {
		return h.handleAgentCard()
	}

	// Handle JSON-RPC A2A requests
	if req.Method == "POST" && strings.Contains(req.Headers["content-type"], "application/json") {
		return h.handleJSONRPC(ctx, req)
	}

	// Default response for unsupported requests
	return h.HandleError("Unsupported request", http.StatusNotFound)
}

// handleCORS handles CORS preflight requests
func (h *Handler) handleCORS() Response {
	return Response{
		Status: http.StatusOK,
		Headers: map[string]string{
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type, Authorization",
			"Access-Control-Max-Age":       "86400",
		},
		Body: "",
	}
}

// handleAgentCard returns the agent card
func (h *Handler) handleAgentCard() Response {
	cardBytes, err := json.Marshal(h.agentCard)
	if err != nil {
		return h.HandleError("Failed to serialize agent card", http.StatusInternalServerError)
	}

	return Response{
		Status: http.StatusOK,
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type, Authorization",
		},
		Body: string(cardBytes),
	}
}

// handleJSONRPC handles JSON-RPC A2A protocol requests
func (h *Handler) handleJSONRPC(ctx context.Context, req Request) Response {
	var jsonrpcReq a2aTypes.JSONRPCRequest
	err := json.Unmarshal([]byte(req.Body), &jsonrpcReq)
	if err != nil {
		return h.handleJSONRPCError(-32700, "Parse error", nil, nil)
	}

	// Validate JSON-RPC request
	err = a2aTypes.ValidateJSONRPCRequest(jsonrpcReq)
	if err != nil {
		return h.handleJSONRPCError(-32600, "Invalid Request", err.Error(), jsonrpcReq.ID)
	}

	// Route to appropriate A2A method
	switch jsonrpcReq.Method {
	case "tasks/get":
		return h.handleGetTask(ctx, jsonrpcReq)
	case "tasks/cancel":
		return h.handleCancelTask(ctx, jsonrpcReq)
	case "message/send":
		return h.handleSendMessage(ctx, jsonrpcReq)
	default:
		return h.handleJSONRPCError(-32601, "Method not found", jsonrpcReq.Method, jsonrpcReq.ID)
	}
}

// handleGetTask handles the tasks/get method
func (h *Handler) handleGetTask(ctx context.Context, req a2aTypes.JSONRPCRequest) Response {
	var params a2a.TaskQueryParams
	if req.Params != nil {
		paramsBytes, _ := json.Marshal(req.Params)
		err := json.Unmarshal(paramsBytes, &params)
		if err != nil {
			return h.handleJSONRPCError(-32602, "Invalid params", err.Error(), req.ID)
		}
	}

	task, err := h.a2aHandler.OnGetTask(ctx, params)
	if err != nil {
		return h.handleJSONRPCError(-32000, "Server error", err.Error(), req.ID)
	}

	return h.handleJSONRPCSuccess(task, req.ID)
}

// handleCancelTask handles the tasks/cancel method
func (h *Handler) handleCancelTask(ctx context.Context, req a2aTypes.JSONRPCRequest) Response {
	var params a2a.TaskIDParams
	if req.Params != nil {
		paramsBytes, _ := json.Marshal(req.Params)
		err := json.Unmarshal(paramsBytes, &params)
		if err != nil {
			return h.handleJSONRPCError(-32602, "Invalid params", err.Error(), req.ID)
		}
	}

	task, err := h.a2aHandler.OnCancelTask(ctx, params)
	if err != nil {
		return h.handleJSONRPCError(-32000, "Server error", err.Error(), req.ID)
	}

	return h.handleJSONRPCSuccess(task, req.ID)
}

// handleSendMessage handles the message/send method
func (h *Handler) handleSendMessage(ctx context.Context, req a2aTypes.JSONRPCRequest) Response {
	var params a2a.MessageSendParams
	if req.Params != nil {
		paramsBytes, _ := json.Marshal(req.Params)
		err := json.Unmarshal(paramsBytes, &params)
		if err != nil {
			return h.handleJSONRPCError(-32602, "Invalid params", err.Error(), req.ID)
		}
	}

	result, err := h.a2aHandler.OnSendMessage(ctx, params)
	if err != nil {
		return h.handleJSONRPCError(-32000, "Server error", err.Error(), req.ID)
	}

	return h.handleJSONRPCSuccess(result, req.ID)
}

// handleJSONRPCSuccess creates a successful JSON-RPC response
func (h *Handler) handleJSONRPCSuccess(result interface{}, id interface{}) Response {
	response := a2aTypes.NewJSONRPCResponse(result, id)
	responseBytes, _ := json.Marshal(response)

	return Response{
		Status: http.StatusOK,
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type, Authorization",
		},
		Body: string(responseBytes),
	}
}

// handleJSONRPCError creates an error JSON-RPC response
func (h *Handler) handleJSONRPCError(code int, message string, data interface{}, id interface{}) Response {
	response := a2aTypes.NewJSONRPCErrorResponse(code, message, data, id)
	responseBytes, _ := json.Marshal(response)

	return Response{
		Status: http.StatusOK, // JSON-RPC errors still return 200 OK
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type, Authorization",
		},
		Body: string(responseBytes),
	}
}

// HandleError creates standardized error responses
func (h *Handler) HandleError(message string, status int) Response {
	errorData := map[string]interface{}{
		"error":     message,
		"timestamp": time.Now().Unix(),
	}

	bodyBytes, _ := json.Marshal(errorData)

	return Response{
		Status: status,
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type, Authorization",
		},
		Body: string(bodyBytes),
	}
}
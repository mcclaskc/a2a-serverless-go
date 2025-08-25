package a2a

import (
	"encoding/json"
	"fmt"
	"strings"
)

// JSON-RPC 2.0 error codes as defined in the specification
const (
	// Standard JSON-RPC error codes
	JSONRPCErrorParseError     = -32700 // Invalid JSON was received by the server
	JSONRPCErrorInvalidRequest = -32600 // The JSON sent is not a valid Request object
	JSONRPCErrorMethodNotFound = -32601 // The method does not exist / is not available
	JSONRPCErrorInvalidParams  = -32602 // Invalid method parameter(s)
	JSONRPCErrorInternalError  = -32603 // Internal JSON-RPC error
	
	// Server error range: -32000 to -32099
	JSONRPCErrorServerError = -32000 // Generic server error
)

// ParseJSONRPCRequest parses raw JSON bytes into a JSONRPCRequest
func ParseJSONRPCRequest(data []byte) (JSONRPCRequest, error) {
	var req JSONRPCRequest
	
	if len(data) == 0 {
		return req, NewJSONRPCParseError("empty request body")
	}
	
	if err := json.Unmarshal(data, &req); err != nil {
		return req, NewJSONRPCParseError(fmt.Sprintf("invalid JSON: %v", err))
	}
	
	if err := ValidateJSONRPCRequest(req); err != nil {
		return req, NewJSONRPCInvalidRequestError(err.Error())
	}
	
	return req, nil
}

// ParseJSONRPCResponse parses raw JSON bytes into a JSONRPCResponse
func ParseJSONRPCResponse(data []byte) (JSONRPCResponse, error) {
	var resp JSONRPCResponse
	
	if len(data) == 0 {
		return resp, NewJSONRPCParseError("empty response body")
	}
	
	if err := json.Unmarshal(data, &resp); err != nil {
		return resp, NewJSONRPCParseError(fmt.Sprintf("invalid JSON: %v", err))
	}
	
	if err := ValidateJSONRPCResponse(resp); err != nil {
		return resp, NewJSONRPCInvalidRequestError(err.Error())
	}
	
	return resp, nil
}

// ValidateJSONRPCResponse validates a JSON-RPC response
func ValidateJSONRPCResponse(resp JSONRPCResponse) error {
	if resp.JSONRPC != "2.0" {
		return fmt.Errorf("jsonrpc must be '2.0'")
	}
	
	// Response must have either result or error, but not both
	hasResult := resp.Result != nil
	hasError := resp.Error != nil
	
	if hasResult && hasError {
		return fmt.Errorf("response cannot have both result and error")
	}
	
	if !hasResult && !hasError {
		return fmt.Errorf("response must have either result or error")
	}
	
	return nil
}

// SerializeJSONRPCRequest serializes a JSONRPCRequest to JSON bytes
func SerializeJSONRPCRequest(req JSONRPCRequest) ([]byte, error) {
	if err := ValidateJSONRPCRequest(req); err != nil {
		return nil, NewJSONRPCInvalidRequestError(err.Error())
	}
	
	data, err := json.Marshal(req)
	if err != nil {
		return nil, NewJSONRPCInternalError(fmt.Sprintf("failed to serialize request: %v", err))
	}
	
	return data, nil
}

// SerializeJSONRPCResponse serializes a JSONRPCResponse to JSON bytes
func SerializeJSONRPCResponse(resp JSONRPCResponse) ([]byte, error) {
	if err := ValidateJSONRPCResponse(resp); err != nil {
		return nil, NewJSONRPCInvalidRequestError(err.Error())
	}
	
	data, err := json.Marshal(resp)
	if err != nil {
		return nil, NewJSONRPCInternalError(fmt.Sprintf("failed to serialize response: %v", err))
	}
	
	return data, nil
}

// IsJSONRPCRequest checks if the given data appears to be a JSON-RPC request
func IsJSONRPCRequest(data []byte) bool {
	// Quick check for JSON-RPC structure without full parsing
	dataStr := strings.TrimSpace(string(data))
	return strings.Contains(dataStr, `"jsonrpc"`) && 
		   strings.Contains(dataStr, `"method"`) &&
		   (strings.Contains(dataStr, `"2.0"`) || strings.Contains(dataStr, "2.0"))
}

// ExtractRequestID attempts to extract the ID from a JSON-RPC request/response
// This is useful for error handling when parsing fails
func ExtractRequestID(data []byte) interface{} {
	var partial struct {
		ID interface{} `json:"id"`
	}
	
	if err := json.Unmarshal(data, &partial); err != nil {
		return nil
	}
	
	return partial.ID
}

// NewJSONRPCParseError creates a JSON-RPC parse error
func NewJSONRPCParseError(message string) *JSONRPCError {
	return &JSONRPCError{
		Code:    JSONRPCErrorParseError,
		Message: "Parse error",
		Data:    message,
	}
}

// NewJSONRPCInvalidRequestError creates a JSON-RPC invalid request error
func NewJSONRPCInvalidRequestError(message string) *JSONRPCError {
	return &JSONRPCError{
		Code:    JSONRPCErrorInvalidRequest,
		Message: "Invalid Request",
		Data:    message,
	}
}

// NewJSONRPCMethodNotFoundError creates a JSON-RPC method not found error
func NewJSONRPCMethodNotFoundError(method string) *JSONRPCError {
	return &JSONRPCError{
		Code:    JSONRPCErrorMethodNotFound,
		Message: "Method not found",
		Data:    fmt.Sprintf("method '%s' not found", method),
	}
}

// NewJSONRPCInvalidParamsError creates a JSON-RPC invalid params error
func NewJSONRPCInvalidParamsError(message string) *JSONRPCError {
	return &JSONRPCError{
		Code:    JSONRPCErrorInvalidParams,
		Message: "Invalid params",
		Data:    message,
	}
}

// NewJSONRPCInternalError creates a JSON-RPC internal error
func NewJSONRPCInternalError(message string) *JSONRPCError {
	return &JSONRPCError{
		Code:    JSONRPCErrorInternalError,
		Message: "Internal error",
		Data:    message,
	}
}

// NewJSONRPCServerError creates a JSON-RPC server error
func NewJSONRPCServerError(code int, message string, data interface{}) *JSONRPCError {
	// Ensure code is in the server error range
	if code > -32000 || code < -32099 {
		code = JSONRPCErrorServerError
	}
	
	return &JSONRPCError{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// HandleJSONRPCError converts a regular error to a JSON-RPC error response
func HandleJSONRPCError(err error, requestID interface{}) JSONRPCResponse {
	if err == nil {
		return NewJSONRPCErrorResponse(
			JSONRPCErrorInternalError,
			"Internal error",
			"nil error passed to HandleJSONRPCError",
			requestID,
		)
	}
	
	// Check if it's already a JSON-RPC error
	if jsonrpcErr, ok := err.(*JSONRPCError); ok {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			Error:   jsonrpcErr,
			ID:      requestID,
		}
	}
	
	// Convert regular error to internal error
	return NewJSONRPCErrorResponse(
		JSONRPCErrorInternalError,
		"Internal error",
		err.Error(),
		requestID,
	)
}

// Error implements the error interface for JSONRPCError
func (e *JSONRPCError) Error() string {
	if e.Data != nil {
		return fmt.Sprintf("JSON-RPC error %d: %s (%v)", e.Code, e.Message, e.Data)
	}
	return fmt.Sprintf("JSON-RPC error %d: %s", e.Code, e.Message)
}
package a2a

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestParseJSONRPCRequest(t *testing.T) {
	tests := []struct {
		name        string
		input       []byte
		expectError bool
		errorType   int // Expected JSON-RPC error code
	}{
		{
			name:        "valid request",
			input:       []byte(`{"jsonrpc":"2.0","method":"test","params":{"key":"value"},"id":1}`),
			expectError: false,
		},
		{
			name:        "valid request with null params",
			input:       []byte(`{"jsonrpc":"2.0","method":"test","params":null,"id":"test-id"}`),
			expectError: false,
		},
		{
			name:        "valid request without params",
			input:       []byte(`{"jsonrpc":"2.0","method":"test","id":1}`),
			expectError: false,
		},
		{
			name:        "empty request body",
			input:       []byte(``),
			expectError: true,
			errorType:   JSONRPCErrorParseError,
		},
		{
			name:        "invalid JSON",
			input:       []byte(`{"jsonrpc":"2.0","method":"test","id":1`),
			expectError: true,
			errorType:   JSONRPCErrorParseError,
		},
		{
			name:        "missing jsonrpc field",
			input:       []byte(`{"method":"test","id":1}`),
			expectError: true,
			errorType:   JSONRPCErrorInvalidRequest,
		},
		{
			name:        "wrong jsonrpc version",
			input:       []byte(`{"jsonrpc":"1.0","method":"test","id":1}`),
			expectError: true,
			errorType:   JSONRPCErrorInvalidRequest,
		},
		{
			name:        "missing method",
			input:       []byte(`{"jsonrpc":"2.0","id":1}`),
			expectError: true,
			errorType:   JSONRPCErrorInvalidRequest,
		},
		{
			name:        "empty method",
			input:       []byte(`{"jsonrpc":"2.0","method":"","id":1}`),
			expectError: true,
			errorType:   JSONRPCErrorInvalidRequest,
		},
		{
			name:        "missing id",
			input:       []byte(`{"jsonrpc":"2.0","method":"test"}`),
			expectError: true,
			errorType:   JSONRPCErrorInvalidRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := ParseJSONRPCRequest(tt.input)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				
				// Check if it's the right type of JSON-RPC error
				if jsonrpcErr, ok := err.(*JSONRPCError); ok {
					if jsonrpcErr.Code != tt.errorType {
						t.Errorf("expected error code %d, got %d", tt.errorType, jsonrpcErr.Code)
					}
				} else {
					t.Errorf("expected JSONRPCError, got %T", err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				
				// Validate the parsed request
				if req.JSONRPC != "2.0" {
					t.Errorf("expected jsonrpc '2.0', got '%s'", req.JSONRPC)
				}
				if req.Method == "" {
					t.Errorf("expected non-empty method")
				}
				if req.ID == nil {
					t.Errorf("expected non-nil ID")
				}
			}
		})
	}
}

func TestParseJSONRPCResponse(t *testing.T) {
	tests := []struct {
		name        string
		input       []byte
		expectError bool
		errorType   int
	}{
		{
			name:        "valid success response",
			input:       []byte(`{"jsonrpc":"2.0","result":{"status":"ok"},"id":1}`),
			expectError: false,
		},
		{
			name:        "valid error response",
			input:       []byte(`{"jsonrpc":"2.0","error":{"code":-32601,"message":"Method not found"},"id":1}`),
			expectError: false,
		},
		{
			name:        "empty response body",
			input:       []byte(``),
			expectError: true,
			errorType:   JSONRPCErrorParseError,
		},
		{
			name:        "invalid JSON",
			input:       []byte(`{"jsonrpc":"2.0","result":{"status":"ok"},"id":1`),
			expectError: true,
			errorType:   JSONRPCErrorParseError,
		},
		{
			name:        "response with both result and error",
			input:       []byte(`{"jsonrpc":"2.0","result":{"status":"ok"},"error":{"code":-32601,"message":"Method not found"},"id":1}`),
			expectError: true,
			errorType:   JSONRPCErrorInvalidRequest,
		},
		{
			name:        "response with neither result nor error",
			input:       []byte(`{"jsonrpc":"2.0","id":1}`),
			expectError: true,
			errorType:   JSONRPCErrorInvalidRequest,
		},
		{
			name:        "wrong jsonrpc version",
			input:       []byte(`{"jsonrpc":"1.0","result":{"status":"ok"},"id":1}`),
			expectError: true,
			errorType:   JSONRPCErrorInvalidRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := ParseJSONRPCResponse(tt.input)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				
				if jsonrpcErr, ok := err.(*JSONRPCError); ok {
					if jsonrpcErr.Code != tt.errorType {
						t.Errorf("expected error code %d, got %d", tt.errorType, jsonrpcErr.Code)
					}
				} else {
					t.Errorf("expected JSONRPCError, got %T", err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				
				if resp.JSONRPC != "2.0" {
					t.Errorf("expected jsonrpc '2.0', got '%s'", resp.JSONRPC)
				}
			}
		})
	}
}

func TestSerializeJSONRPCRequest(t *testing.T) {
	tests := []struct {
		name        string
		request     JSONRPCRequest
		expectError bool
	}{
		{
			name: "valid request",
			request: JSONRPCRequest{
				JSONRPC: "2.0",
				Method:  "test",
				Params:  map[string]string{"key": "value"},
				ID:      1,
			},
			expectError: false,
		},
		{
			name: "invalid request - missing jsonrpc",
			request: JSONRPCRequest{
				Method: "test",
				ID:     1,
			},
			expectError: true,
		},
		{
			name: "invalid request - empty method",
			request: JSONRPCRequest{
				JSONRPC: "2.0",
				Method:  "",
				ID:      1,
			},
			expectError: true,
		},
		{
			name: "invalid request - nil ID",
			request: JSONRPCRequest{
				JSONRPC: "2.0",
				Method:  "test",
				ID:      nil,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := SerializeJSONRPCRequest(tt.request)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				
				// Verify we can parse it back
				var parsed JSONRPCRequest
				if err := json.Unmarshal(data, &parsed); err != nil {
					t.Errorf("failed to parse serialized request: %v", err)
				}
			}
		})
	}
}

func TestSerializeJSONRPCResponse(t *testing.T) {
	tests := []struct {
		name        string
		response    JSONRPCResponse
		expectError bool
	}{
		{
			name: "valid success response",
			response: JSONRPCResponse{
				JSONRPC: "2.0",
				Result:  map[string]string{"status": "ok"},
				ID:      1,
			},
			expectError: false,
		},
		{
			name: "valid error response",
			response: JSONRPCResponse{
				JSONRPC: "2.0",
				Error: &JSONRPCError{
					Code:    -32601,
					Message: "Method not found",
				},
				ID: 1,
			},
			expectError: false,
		},
		{
			name: "invalid response - both result and error",
			response: JSONRPCResponse{
				JSONRPC: "2.0",
				Result:  map[string]string{"status": "ok"},
				Error: &JSONRPCError{
					Code:    -32601,
					Message: "Method not found",
				},
				ID: 1,
			},
			expectError: true,
		},
		{
			name: "invalid response - neither result nor error",
			response: JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      1,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := SerializeJSONRPCResponse(tt.response)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				
				// Verify we can parse it back
				var parsed JSONRPCResponse
				if err := json.Unmarshal(data, &parsed); err != nil {
					t.Errorf("failed to parse serialized response: %v", err)
				}
			}
		})
	}
}

func TestIsJSONRPCRequest(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected bool
	}{
		{
			name:     "valid JSON-RPC request",
			input:    []byte(`{"jsonrpc":"2.0","method":"test","id":1}`),
			expected: true,
		},
		{
			name:     "valid JSON-RPC request with spacing",
			input:    []byte(`  {"jsonrpc": "2.0", "method": "test", "id": 1}  `),
			expected: true,
		},
		{
			name:     "JSON-RPC request with numeric version",
			input:    []byte(`{"jsonrpc":2.0,"method":"test","id":1}`),
			expected: true,
		},
		{
			name:     "regular JSON object",
			input:    []byte(`{"key":"value","number":123}`),
			expected: false,
		},
		{
			name:     "empty JSON object",
			input:    []byte(`{}`),
			expected: false,
		},
		{
			name:     "invalid JSON",
			input:    []byte(`{"key":"value"`),
			expected: false,
		},
		{
			name:     "empty input",
			input:    []byte(``),
			expected: false,
		},
		{
			name:     "JSON-RPC response",
			input:    []byte(`{"jsonrpc":"2.0","result":{"status":"ok"},"id":1}`),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsJSONRPCRequest(tt.input)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestExtractRequestID(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected interface{}
	}{
		{
			name:     "numeric ID",
			input:    []byte(`{"jsonrpc":"2.0","method":"test","id":123}`),
			expected: float64(123), // JSON unmarshaling converts numbers to float64
		},
		{
			name:     "string ID",
			input:    []byte(`{"jsonrpc":"2.0","method":"test","id":"test-id"}`),
			expected: "test-id",
		},
		{
			name:     "null ID",
			input:    []byte(`{"jsonrpc":"2.0","method":"test","id":null}`),
			expected: nil,
		},
		{
			name:     "missing ID",
			input:    []byte(`{"jsonrpc":"2.0","method":"test"}`),
			expected: nil,
		},
		{
			name:     "invalid JSON",
			input:    []byte(`{"jsonrpc":"2.0","method":"test","id":123`),
			expected: nil,
		},
		{
			name:     "empty input",
			input:    []byte(``),
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractRequestID(tt.input)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestJSONRPCErrorCreation(t *testing.T) {
	tests := []struct {
		name         string
		errorFunc    func() *JSONRPCError
		expectedCode int
	}{
		{
			name:         "parse error",
			errorFunc:    func() *JSONRPCError { return NewJSONRPCParseError("test message") },
			expectedCode: JSONRPCErrorParseError,
		},
		{
			name:         "invalid request error",
			errorFunc:    func() *JSONRPCError { return NewJSONRPCInvalidRequestError("test message") },
			expectedCode: JSONRPCErrorInvalidRequest,
		},
		{
			name:         "method not found error",
			errorFunc:    func() *JSONRPCError { return NewJSONRPCMethodNotFoundError("test_method") },
			expectedCode: JSONRPCErrorMethodNotFound,
		},
		{
			name:         "invalid params error",
			errorFunc:    func() *JSONRPCError { return NewJSONRPCInvalidParamsError("test message") },
			expectedCode: JSONRPCErrorInvalidParams,
		},
		{
			name:         "internal error",
			errorFunc:    func() *JSONRPCError { return NewJSONRPCInternalError("test message") },
			expectedCode: JSONRPCErrorInternalError,
		},
		{
			name:         "server error",
			errorFunc:    func() *JSONRPCError { return NewJSONRPCServerError(-32001, "test message", nil) },
			expectedCode: -32001,
		},
		{
			name:         "server error with invalid code",
			errorFunc:    func() *JSONRPCError { return NewJSONRPCServerError(-1000, "test message", nil) },
			expectedCode: JSONRPCErrorServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.errorFunc()
			if err.Code != tt.expectedCode {
				t.Errorf("expected error code %d, got %d", tt.expectedCode, err.Code)
			}
			if err.Message == "" {
				t.Errorf("expected non-empty error message")
			}
		})
	}
}

func TestHandleJSONRPCError(t *testing.T) {
	tests := []struct {
		name        string
		inputError  error
		requestID   interface{}
		expectedCode int
	}{
		{
			name:         "nil error",
			inputError:   nil,
			requestID:    1,
			expectedCode: JSONRPCErrorInternalError,
		},
		{
			name:         "JSON-RPC error",
			inputError:   NewJSONRPCMethodNotFoundError("test"),
			requestID:    "test-id",
			expectedCode: JSONRPCErrorMethodNotFound,
		},
		{
			name:         "regular error",
			inputError:   errors.New("test error"),
			requestID:    123,
			expectedCode: JSONRPCErrorInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := HandleJSONRPCError(tt.inputError, tt.requestID)
			
			if resp.JSONRPC != "2.0" {
				t.Errorf("expected jsonrpc '2.0', got '%s'", resp.JSONRPC)
			}
			if resp.Error == nil {
				t.Errorf("expected error in response")
				return
			}
			if resp.Error.Code != tt.expectedCode {
				t.Errorf("expected error code %d, got %d", tt.expectedCode, resp.Error.Code)
			}
			if resp.ID != tt.requestID {
				t.Errorf("expected ID %v, got %v", tt.requestID, resp.ID)
			}
			if resp.Result != nil {
				t.Errorf("expected nil result in error response")
			}
		})
	}
}

func TestJSONRPCErrorInterface(t *testing.T) {
	err := NewJSONRPCMethodNotFoundError("test_method")
	
	// Test that it implements the error interface
	var _ error = err
	
	// Test the Error() method
	errStr := err.Error()
	if errStr == "" {
		t.Errorf("expected non-empty error string")
	}
	
	// Should contain the error code and message
	if !contains(errStr, "JSON-RPC error") {
		t.Errorf("error string should contain 'JSON-RPC error'")
	}
	if !contains(errStr, "-32601") {
		t.Errorf("error string should contain the error code")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || 
		s[len(s)-len(substr):] == substr || 
		containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
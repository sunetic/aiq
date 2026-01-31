package mocks

import (
	"context"

	"github.com/aiq/aiq/internal/llm"
)

// MockLLMClient is a mock implementation of the LLM client for testing
type MockLLMClient struct {
	// ChatWithToolsFunc allows setting a custom function for ChatWithTools
	ChatWithToolsFunc func(ctx context.Context, messages []interface{}, tools []llm.Function) (*llm.ChatResponse, error)

	// Default response to return if ChatWithToolsFunc is not set
	DefaultResponse *llm.ChatResponse
	DefaultError    error

	// Track calls for verification
	Calls []ChatWithToolsCall
}

// ChatWithToolsCall records a call to ChatWithTools
type ChatWithToolsCall struct {
	Context  context.Context
	Messages []interface{}
	Tools    []llm.Function
}

// ChatWithTools implements the LLM client interface
func (m *MockLLMClient) ChatWithTools(ctx context.Context, messages []interface{}, tools []llm.Function) (*llm.ChatResponse, error) {
	// Record the call
	m.Calls = append(m.Calls, ChatWithToolsCall{
		Context:  ctx,
		Messages: messages,
		Tools:    tools,
	})

	// Use custom function if provided
	if m.ChatWithToolsFunc != nil {
		return m.ChatWithToolsFunc(ctx, messages, tools)
	}

	// Return default response or error
	return m.DefaultResponse, m.DefaultError
}

// Reset clears all recorded calls
func (m *MockLLMClient) Reset() {
	m.Calls = nil
	m.ChatWithToolsFunc = nil
	m.DefaultResponse = nil
	m.DefaultError = nil
}

// NewMockLLMClient creates a new mock LLM client
func NewMockLLMClient() *MockLLMClient {
	return &MockLLMClient{
		Calls: make([]ChatWithToolsCall, 0),
	}
}

// WithResponse sets a default response for the mock
func (m *MockLLMClient) WithResponse(response *llm.ChatResponse) *MockLLMClient {
	m.DefaultResponse = response
	return m
}

// WithError sets a default error for the mock
func (m *MockLLMClient) WithError(err error) *MockLLMClient {
	m.DefaultError = err
	return m
}

// WithChatWithToolsFunc sets a custom function for ChatWithTools
func (m *MockLLMClient) WithChatWithToolsFunc(fn func(ctx context.Context, messages []interface{}, tools []llm.Function) (*llm.ChatResponse, error)) *MockLLMClient {
	m.ChatWithToolsFunc = fn
	return m
}

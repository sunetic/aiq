package mocks

// MockUI provides mock implementations for UI components for testing
// This allows testing code that uses UI functions without requiring actual terminal interaction
type MockUI struct {
	// ShowConfirmFunc allows setting a custom function for ShowConfirm
	ShowConfirmFunc func(message string) (bool, error)

	// Default response for ShowConfirm
	DefaultConfirm bool
	DefaultError   error

	// Track calls for verification
	ShowConfirmCalls []ShowConfirmCall
}

// ShowConfirmCall records a call to ShowConfirm
type ShowConfirmCall struct {
	Message string
}

// ShowConfirm implements a mock version of ui.ShowConfirm
func (m *MockUI) ShowConfirm(message string) (bool, error) {
	// Record the call
	m.ShowConfirmCalls = append(m.ShowConfirmCalls, ShowConfirmCall{
		Message: message,
	})

	// Use custom function if provided
	if m.ShowConfirmFunc != nil {
		return m.ShowConfirmFunc(message)
	}

	// Return default response
	return m.DefaultConfirm, m.DefaultError
}

// Reset clears all recorded calls
func (m *MockUI) Reset() {
	m.ShowConfirmCalls = nil
	m.ShowConfirmFunc = nil
	m.DefaultConfirm = false
	m.DefaultError = nil
}

// NewMockUI creates a new mock UI
func NewMockUI() *MockUI {
	return &MockUI{
		ShowConfirmCalls: make([]ShowConfirmCall, 0),
	}
}

// WithConfirm sets default confirm response
func (m *MockUI) WithConfirm(confirm bool) *MockUI {
	m.DefaultConfirm = confirm
	return m
}

// WithError sets default error response
func (m *MockUI) WithError(err error) *MockUI {
	m.DefaultError = err
	return m
}

// WithShowConfirmFunc sets a custom function for ShowConfirm
func (m *MockUI) WithShowConfirmFunc(fn func(message string) (bool, error)) *MockUI {
	m.ShowConfirmFunc = fn
	return m
}

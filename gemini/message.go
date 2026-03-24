package gemini

import "encoding/json"

// Message is the interface implemented by all message types streamed
// from the Gemini CLI.
type Message interface {
	Type() string
}

// InitMessage is emitted when the CLI initializes a session.
type InitMessage struct {
	SessionID string `json:"session_id,omitempty"`
	Model     string `json:"model,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
}

func (*InitMessage) Type() string { return "init" }

// AssistantMessage is a response from the model.
type AssistantMessage struct {
	Content   string `json:"content"`
	Timestamp string `json:"timestamp,omitempty"`
	Delta     bool   `json:"delta,omitempty"`
}

func (*AssistantMessage) Type() string { return "assistant" }

// UserMessage represents a user turn in the conversation.
type UserMessage struct {
	Content   string `json:"content"`
	Timestamp string `json:"timestamp,omitempty"`
}

func (*UserMessage) Type() string { return "user" }

// ToolUseMessage represents a tool invocation by the model.
type ToolUseMessage struct {
	ToolName   string          `json:"tool_name"`
	ToolID     string          `json:"tool_id"`
	Parameters json.RawMessage `json:"parameters"`
	Timestamp  string          `json:"timestamp,omitempty"`
}

func (*ToolUseMessage) Type() string { return "tool_use" }

// ToolResultMessage represents the result of a tool invocation.
type ToolResultMessage struct {
	ToolID    string     `json:"tool_id"`
	Status    string     `json:"status"` // "success" or "error"
	Output    string     `json:"output"`
	Error     *ToolError `json:"error,omitempty"`
	Timestamp string     `json:"timestamp,omitempty"`
}

func (*ToolResultMessage) Type() string { return "tool_result" }

// ToolError describes a tool execution failure.
type ToolError struct {
	// ErrorType is the category of error (JSON field: "type").
	ErrorType string `json:"type"`
	Message   string `json:"message"`
}

// ErrorMessage represents a CLI-level error event.
// Emitted when the CLI encounters an unrecoverable error.
type ErrorMessage struct {
	ErrorText string `json:"message"`
	Timestamp string `json:"timestamp,omitempty"`
}

func (*ErrorMessage) Type() string { return "error" }

// ResultMessage is the final event with usage statistics.
type ResultMessage struct {
	Status    string      `json:"status"` // "success" or "error"
	Stats     ResultStats `json:"stats"`
	Timestamp string      `json:"timestamp,omitempty"`
}

func (*ResultMessage) Type() string { return "result" }

// ResultStats holds usage statistics from a completed run.
type ResultStats struct {
	TotalTokens  int                   `json:"total_tokens"`
	InputTokens  int                   `json:"input_tokens"`
	OutputTokens int                   `json:"output_tokens"`
	Cached       int                   `json:"cached"`
	Input        int                   `json:"input"` // non-cached input tokens
	DurationMs   float64               `json:"duration_ms"`
	ToolCalls    int                   `json:"tool_calls"`
	Models       map[string]ModelStats `json:"models,omitempty"`
}

// ModelStats holds per-model token usage.
type ModelStats struct {
	TotalTokens  int `json:"total_tokens"`
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	Cached       int `json:"cached"`
	Input        int `json:"input"`
}

// UnknownMessage wraps any event type the SDK does not recognize.
// This ensures forward compatibility — new event types from future
// CLI versions will not break existing code.
type UnknownMessage struct {
	RawType string
	Raw     json.RawMessage
}

func (m *UnknownMessage) Type() string { return m.RawType }

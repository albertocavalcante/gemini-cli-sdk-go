package gemini

import "testing"

// Compile-time interface compliance checks.
var (
	_ Message = (*InitMessage)(nil)
	_ Message = (*AssistantMessage)(nil)
	_ Message = (*UserMessage)(nil)
	_ Message = (*ToolUseMessage)(nil)
	_ Message = (*ToolResultMessage)(nil)
	_ Message = (*ErrorMessage)(nil)
	_ Message = (*ResultMessage)(nil)
	_ Message = (*UnknownMessage)(nil)
)

// Compile-time error interface checks.
var (
	_ error = (*CLIError)(nil)
	_ error = (*ProtocolError)(nil)
	_ error = (*ProcessError)(nil)
)

func TestMessageTypes(t *testing.T) {
	tests := []struct {
		msg  Message
		want string
	}{
		{&InitMessage{}, "init"},
		{&AssistantMessage{}, "assistant"},
		{&UserMessage{}, "user"},
		{&ToolUseMessage{}, "tool_use"},
		{&ToolResultMessage{}, "tool_result"},
		{&ErrorMessage{}, "error"},
		{&ResultMessage{}, "result"},
		{&UnknownMessage{RawType: "custom"}, "custom"},
	}
	for _, tt := range tests {
		if got := tt.msg.Type(); got != tt.want {
			t.Errorf("%T.Type() = %q, want %q", tt.msg, got, tt.want)
		}
	}
}

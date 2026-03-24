package gemini

import (
	"context"
	"encoding/json"
	"regexp"
)

// HookEvent identifies when a hook fires.
type HookEvent string

const (
	HookPreToolUse  HookEvent = "PreToolUse"
	HookPostToolUse HookEvent = "PostToolUse"
	HookMessage     HookEvent = "Message"
	HookResult      HookEvent = "Result"
)

// HookRegistration binds a callback to an event.
type HookRegistration struct {
	Event       HookEvent
	ToolPattern string // Optional regex to match tool names.
	Callback    HookCallback
}

// HookCallback is invoked when a matching hook event fires.
type HookCallback func(ctx context.Context, event HookInput) (HookOutput, error)

// HookInput provides context to hook callbacks.
type HookInput struct {
	Event      HookEvent
	SessionID  string
	Message    Message
	ToolName   string
	ToolInput  json.RawMessage
	ToolOutput string
}

// HookOutput controls hook behavior.
type HookOutput struct {
	Block  bool   // If true, block the tool call (PreToolUse only).
	Reason string // Explanation for blocking.
}

// hookRunner manages lifecycle hook execution during a session.
type hookRunner struct {
	hooks     []HookRegistration
	sessionID string
	// toolNames maps tool_id → tool_name for PostToolUse dispatch.
	toolNames map[string]string
}

func newHookRunner(hooks []HookRegistration, sessionID string) *hookRunner {
	return &hookRunner{
		hooks:     hooks,
		sessionID: sessionID,
		toolNames: make(map[string]string),
	}
}

// fireHooks dispatches hooks for a received message.
func (hr *hookRunner) fireHooks(ctx context.Context, msg Message) {
	// Track tool names for PostToolUse lookup.
	if tu, ok := msg.(*ToolUseMessage); ok {
		hr.toolNames[tu.ToolID] = tu.ToolName
	}

	for _, h := range hr.hooks {
		switch h.Event {
		case HookMessage:
			hr.invoke(ctx, h, HookInput{
				Event:     HookMessage,
				SessionID: hr.sessionID,
				Message:   msg,
			})

		case HookPreToolUse:
			tu, ok := msg.(*ToolUseMessage)
			if !ok || !matchPattern(h.ToolPattern, tu.ToolName) {
				continue
			}
			hr.invoke(ctx, h, HookInput{
				Event:     HookPreToolUse,
				SessionID: hr.sessionID,
				Message:   msg,
				ToolName:  tu.ToolName,
				ToolInput: tu.Parameters,
			})

		case HookPostToolUse:
			tr, ok := msg.(*ToolResultMessage)
			if !ok {
				continue
			}
			toolName := hr.toolNames[tr.ToolID]
			if !matchPattern(h.ToolPattern, toolName) {
				continue
			}
			hr.invoke(ctx, h, HookInput{
				Event:      HookPostToolUse,
				SessionID:  hr.sessionID,
				Message:    msg,
				ToolName:   toolName,
				ToolOutput: tr.Output,
			})

		case HookResult:
			if _, ok := msg.(*ResultMessage); !ok {
				continue
			}
			hr.invoke(ctx, h, HookInput{
				Event:     HookResult,
				SessionID: hr.sessionID,
				Message:   msg,
			})
		}
	}
}

// invoke calls a hook callback. Errors are silently ignored to avoid
// breaking the message stream.
func (*hookRunner) invoke(ctx context.Context, h HookRegistration, input HookInput) {
	_, _ = h.Callback(ctx, input)
}

// matchPattern returns true if toolName matches the regex pattern.
// Empty pattern matches everything. Invalid regex does not match.
func matchPattern(pattern, toolName string) bool {
	if pattern == "" {
		return true
	}
	matched, err := regexp.MatchString(pattern, toolName)
	return err == nil && matched
}

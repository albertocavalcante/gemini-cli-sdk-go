package gemini

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	"github.com/albertocavalcante/gemini-cli-sdk-go/internal/transport"
)

func TestNewClient(t *testing.T) {
	c := NewClient(Options{Model: ModelFlash})
	if c == nil {
		t.Fatal("NewClient returned nil")
	}
	if c.SessionID() != "" {
		t.Errorf("initial SessionID = %q, want empty", c.SessionID())
	}
}

func TestClientSessionIDFromInit(t *testing.T) {
	mock := &transport.MockTransport{
		RawLines: []json.RawMessage{
			json.RawMessage(`{"type":"init","session_id":"captured_id","model":"flash"}`),
			json.RawMessage(`{"type":"result","status":"success","stats":{}}`),
		},
	}

	c := newClientWithTransport(Options{}, mock)
	for msg := range c.Query(context.Background(), "hello") {
		if msg.Err != nil {
			t.Fatalf("error: %v", msg.Err)
		}
	}

	if c.SessionID() != "captured_id" {
		t.Errorf("SessionID = %q, want %q", c.SessionID(), "captured_id")
	}
}

func TestClientHookMessage(t *testing.T) {
	var mu sync.Mutex
	var received []string

	mock := &transport.MockTransport{
		RawLines: []json.RawMessage{
			json.RawMessage(`{"type":"init","session_id":"s1"}`),
			json.RawMessage(`{"type":"message","role":"assistant","content":"hi"}`),
			json.RawMessage(`{"type":"result","status":"success","stats":{}}`),
		},
	}

	c := newClientWithTransport(Options{
		Hooks: []HookRegistration{{
			Event: HookMessage,
			Callback: func(_ context.Context, ev HookInput) (HookOutput, error) {
				mu.Lock()
				received = append(received, ev.Message.Type())
				mu.Unlock()
				return HookOutput{}, nil
			},
		}},
	}, mock)

	for range c.Query(context.Background(), "hello") {
	}

	mu.Lock()
	defer mu.Unlock()
	if len(received) != 3 {
		t.Errorf("hook fired %d times, want 3: %v", len(received), received)
	}
}

func TestClientHookPreToolUse(t *testing.T) {
	var toolName string

	mock := &transport.MockTransport{
		RawLines: []json.RawMessage{
			json.RawMessage(`{"type":"tool_use","tool_name":"grep_search","tool_id":"t1","parameters":{"query":"foo"}}`),
			json.RawMessage(`{"type":"tool_result","tool_id":"t1","status":"success","output":"found"}`),
			json.RawMessage(`{"type":"result","status":"success","stats":{}}`),
		},
	}

	c := newClientWithTransport(Options{
		Hooks: []HookRegistration{{
			Event: HookPreToolUse,
			Callback: func(_ context.Context, ev HookInput) (HookOutput, error) {
				toolName = ev.ToolName
				return HookOutput{}, nil
			},
		}},
	}, mock)

	for range c.Query(context.Background(), "search") {
	}

	if toolName != "grep_search" {
		t.Errorf("toolName = %q, want %q", toolName, "grep_search")
	}
}

func TestClientHookPostToolUse(t *testing.T) {
	var output string

	mock := &transport.MockTransport{
		RawLines: []json.RawMessage{
			json.RawMessage(`{"type":"tool_use","tool_name":"read_file","tool_id":"t1","parameters":{}}`),
			json.RawMessage(`{"type":"tool_result","tool_id":"t1","status":"success","output":"content here"}`),
			json.RawMessage(`{"type":"result","status":"success","stats":{}}`),
		},
	}

	c := newClientWithTransport(Options{
		Hooks: []HookRegistration{{
			Event: HookPostToolUse,
			Callback: func(_ context.Context, ev HookInput) (HookOutput, error) {
				output = ev.ToolOutput
				return HookOutput{}, nil
			},
		}},
	}, mock)

	for range c.Query(context.Background(), "read") {
	}

	if output != "content here" {
		t.Errorf("output = %q, want %q", output, "content here")
	}
}

func TestClientHookToolPattern(t *testing.T) {
	var matched bool

	mock := &transport.MockTransport{
		RawLines: []json.RawMessage{
			json.RawMessage(`{"type":"tool_use","tool_name":"read_file","tool_id":"t1","parameters":{}}`),
			json.RawMessage(`{"type":"tool_result","tool_id":"t1","status":"success","output":""}`),
			json.RawMessage(`{"type":"tool_use","tool_name":"write_file","tool_id":"t2","parameters":{}}`),
			json.RawMessage(`{"type":"tool_result","tool_id":"t2","status":"success","output":""}`),
			json.RawMessage(`{"type":"result","status":"success","stats":{}}`),
		},
	}

	c := newClientWithTransport(Options{
		Hooks: []HookRegistration{{
			Event:       HookPreToolUse,
			ToolPattern: "^write",
			Callback: func(_ context.Context, _ HookInput) (HookOutput, error) {
				matched = true
				return HookOutput{}, nil
			},
		}},
	}, mock)

	for range c.Query(context.Background(), "test") {
	}

	if !matched {
		t.Error("hook with pattern ^write should match write_file")
	}
}

func TestClientHookResult(t *testing.T) {
	var status string

	mock := &transport.MockTransport{
		RawLines: []json.RawMessage{
			json.RawMessage(`{"type":"result","status":"success","stats":{"total_tokens":42}}`),
		},
	}

	c := newClientWithTransport(Options{
		Hooks: []HookRegistration{{
			Event: HookResult,
			Callback: func(_ context.Context, ev HookInput) (HookOutput, error) {
				if rm, ok := ev.Message.(*ResultMessage); ok {
					status = rm.Status
				}
				return HookOutput{}, nil
			},
		}},
	}, mock)

	for range c.Query(context.Background(), "test") {
	}

	if status != "success" {
		t.Errorf("status = %q, want %q", status, "success")
	}
}

func TestClientHookErrorDoesNotBreakStream(t *testing.T) {
	mock := &transport.MockTransport{
		RawLines: []json.RawMessage{
			json.RawMessage(`{"type":"message","role":"assistant","content":"hello"}`),
			json.RawMessage(`{"type":"result","status":"success","stats":{}}`),
		},
	}

	c := newClientWithTransport(Options{
		Hooks: []HookRegistration{{
			Event: HookMessage,
			Callback: func(_ context.Context, _ HookInput) (HookOutput, error) {
				return HookOutput{}, context.Canceled // simulate error
			},
		}},
	}, mock)

	var count int
	for msg := range c.Query(context.Background(), "test") {
		if msg.Err != nil {
			t.Fatalf("hook error should not propagate: %v", msg.Err)
		}
		count++
	}
	if count != 2 {
		t.Errorf("received %d messages, want 2", count)
	}
}

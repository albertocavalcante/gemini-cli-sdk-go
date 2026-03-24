package gemini

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/albertocavalcante/gemini-cli-sdk-go/internal/transport"
)

func TestQueryWithMockTransport(t *testing.T) {
	mock := &transport.MockTransport{
		RawLines: []json.RawMessage{
			json.RawMessage(`{"type":"init","session_id":"s1","model":"flash"}`),
			json.RawMessage(`{"type":"message","role":"assistant","content":"Hi!"}`),
			json.RawMessage(`{"type":"result","status":"success","stats":{"total_tokens":10}}`),
		},
	}

	var types []string
	for msg := range queryWithTransport(context.Background(), "hello", Options{}, mock) {
		if msg.Err != nil {
			t.Fatalf("unexpected error: %v", msg.Err)
		}
		types = append(types, msg.Message.Type())
	}

	want := []string{"init", "assistant", "result"}
	if len(types) != len(want) {
		t.Fatalf("got %d messages, want %d: %v", len(types), len(want), types)
	}
	for i, w := range want {
		if types[i] != w {
			t.Errorf("types[%d] = %q, want %q", i, types[i], w)
		}
	}
}

func TestQueryWithStartError(t *testing.T) {
	mock := &transport.MockTransport{StartErr: errors.New("failed")}

	var errs []error
	for msg := range queryWithTransport(context.Background(), "hello", Options{}, mock) {
		if msg.Err != nil {
			errs = append(errs, msg.Err)
		}
	}
	if len(errs) == 0 {
		t.Fatal("expected at least one error")
	}
}

func TestQueryWithContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	mock := &transport.MockTransport{
		RawLines: []json.RawMessage{
			json.RawMessage(`{"type":"init","session_id":"s1"}`),
			json.RawMessage(`{"type":"message","role":"assistant","content":"never seen"}`),
		},
	}

	var count int
	for range queryWithTransport(ctx, "hello", Options{}, mock) {
		count++
	}
	// Should receive very few messages due to cancellation.
	// The exact count depends on goroutine scheduling.
	_ = count
}

func TestQueryWithUnknownMessageType(t *testing.T) {
	mock := &transport.MockTransport{
		RawLines: []json.RawMessage{
			json.RawMessage(`{"type":"new_future_type","data":"whatever"}`),
		},
	}

	for msg := range queryWithTransport(context.Background(), "hello", Options{}, mock) {
		if msg.Err != nil {
			t.Fatalf("unknown types should not error: %v", msg.Err)
		}
		if msg.Message.Type() != "new_future_type" {
			t.Errorf("type = %q, want %q", msg.Message.Type(), "new_future_type")
		}
	}
}

func TestQueryWithToolUseFlow(t *testing.T) {
	mock := &transport.MockTransport{
		RawLines: []json.RawMessage{
			json.RawMessage(`{"type":"init","session_id":"s1"}`),
			json.RawMessage(`{"type":"tool_use","tool_name":"read_file","tool_id":"t1","parameters":{"path":"go.mod"}}`),
			json.RawMessage(`{"type":"tool_result","tool_id":"t1","status":"success","output":"module foo"}`),
			json.RawMessage(`{"type":"message","role":"assistant","content":"The module is foo."}`),
			json.RawMessage(`{"type":"result","status":"success","stats":{"total_tokens":50,"tool_calls":1}}`),
		},
	}

	var types []string
	for msg := range queryWithTransport(context.Background(), "read go.mod", Options{}, mock) {
		if msg.Err != nil {
			t.Fatalf("unexpected error: %v", msg.Err)
		}
		types = append(types, msg.Message.Type())
	}

	want := []string{"init", "tool_use", "tool_result", "assistant", "result"}
	if len(types) != len(want) {
		t.Fatalf("got %v, want %v", types, want)
	}
	for i, w := range want {
		if types[i] != w {
			t.Errorf("types[%d] = %q, want %q", i, types[i], w)
		}
	}
}

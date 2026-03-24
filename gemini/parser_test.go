package gemini

import (
	"testing"
)

func TestParseInitMessage(t *testing.T) {
	data := []byte(`{"type":"init","session_id":"sess_abc","model":"gemini-2.5-pro","timestamp":"2025-01-01T00:00:00Z"}`)
	msg, err := ParseMessage(data)
	if err != nil {
		t.Fatalf("ParseMessage: %v", err)
	}
	im, ok := msg.(*InitMessage)
	if !ok {
		t.Fatalf("got %T, want *InitMessage", msg)
	}
	if im.SessionID != "sess_abc" {
		t.Errorf("SessionID = %q, want %q", im.SessionID, "sess_abc")
	}
	if im.Model != "gemini-2.5-pro" {
		t.Errorf("Model = %q, want %q", im.Model, "gemini-2.5-pro")
	}
}

func TestParseAssistantMessage(t *testing.T) {
	data := []byte(`{"type":"message","role":"assistant","content":"Hello!","delta":true}`)
	msg, err := ParseMessage(data)
	if err != nil {
		t.Fatalf("ParseMessage: %v", err)
	}
	am, ok := msg.(*AssistantMessage)
	if !ok {
		t.Fatalf("got %T, want *AssistantMessage", msg)
	}
	if am.Content != "Hello!" {
		t.Errorf("Content = %q, want %q", am.Content, "Hello!")
	}
	if !am.Delta {
		t.Error("Delta = false, want true")
	}
}

func TestParseUserMessage(t *testing.T) {
	data := []byte(`{"type":"message","role":"user","content":"What is Go?"}`)
	msg, err := ParseMessage(data)
	if err != nil {
		t.Fatalf("ParseMessage: %v", err)
	}
	um, ok := msg.(*UserMessage)
	if !ok {
		t.Fatalf("got %T, want *UserMessage", msg)
	}
	if um.Content != "What is Go?" {
		t.Errorf("Content = %q", um.Content)
	}
}

func TestParseToolUseMessage(t *testing.T) {
	data := []byte(`{"type":"tool_use","tool_name":"read_file","tool_id":"t1","parameters":{"path":"main.go"}}`)
	msg, err := ParseMessage(data)
	if err != nil {
		t.Fatalf("ParseMessage: %v", err)
	}
	tu, ok := msg.(*ToolUseMessage)
	if !ok {
		t.Fatalf("got %T, want *ToolUseMessage", msg)
	}
	if tu.ToolName != "read_file" {
		t.Errorf("ToolName = %q", tu.ToolName)
	}
	if tu.ToolID != "t1" {
		t.Errorf("ToolID = %q", tu.ToolID)
	}
	if string(tu.Parameters) != `{"path":"main.go"}` {
		t.Errorf("Parameters = %s", tu.Parameters)
	}
}

func TestParseToolResultMessage(t *testing.T) {
	data := []byte(`{"type":"tool_result","tool_id":"t1","status":"success","output":"file contents"}`)
	msg, err := ParseMessage(data)
	if err != nil {
		t.Fatalf("ParseMessage: %v", err)
	}
	tr, ok := msg.(*ToolResultMessage)
	if !ok {
		t.Fatalf("got %T, want *ToolResultMessage", msg)
	}
	if tr.Status != "success" {
		t.Errorf("Status = %q", tr.Status)
	}
	if tr.Output != "file contents" {
		t.Errorf("Output = %q", tr.Output)
	}
}

func TestParseToolResultWithError(t *testing.T) {
	data := []byte(`{"type":"tool_result","tool_id":"t2","status":"error","output":"","error":{"type":"not_found","message":"file not found"}}`)
	msg, err := ParseMessage(data)
	if err != nil {
		t.Fatalf("ParseMessage: %v", err)
	}
	tr := msg.(*ToolResultMessage)
	if tr.Error == nil {
		t.Fatal("Error should be non-nil")
	}
	if tr.Error.Message != "file not found" {
		t.Errorf("Error.Message = %q", tr.Error.Message)
	}
}

func TestParseResultMessage(t *testing.T) {
	data := []byte(`{"type":"result","status":"success","stats":{"total_tokens":100,"input_tokens":50,"output_tokens":50,"duration_ms":1234.5,"tool_calls":2}}`)
	msg, err := ParseMessage(data)
	if err != nil {
		t.Fatalf("ParseMessage: %v", err)
	}
	rm, ok := msg.(*ResultMessage)
	if !ok {
		t.Fatalf("got %T, want *ResultMessage", msg)
	}
	if rm.Status != "success" {
		t.Errorf("Status = %q", rm.Status)
	}
	if rm.Stats.TotalTokens != 100 {
		t.Errorf("TotalTokens = %d", rm.Stats.TotalTokens)
	}
	if rm.Stats.ToolCalls != 2 {
		t.Errorf("ToolCalls = %d", rm.Stats.ToolCalls)
	}
}

func TestParseUnknownType(t *testing.T) {
	data := []byte(`{"type":"future_event","data":"something"}`)
	msg, err := ParseMessage(data)
	if err != nil {
		t.Fatalf("ParseMessage: %v", err)
	}
	um, ok := msg.(*UnknownMessage)
	if !ok {
		t.Fatalf("got %T, want *UnknownMessage", msg)
	}
	if um.RawType != "future_event" {
		t.Errorf("RawType = %q", um.RawType)
	}
}

func TestParseUnknownRole(t *testing.T) {
	data := []byte(`{"type":"message","role":"system","content":"hello"}`)
	msg, err := ParseMessage(data)
	if err != nil {
		t.Fatalf("ParseMessage: %v", err)
	}
	um, ok := msg.(*UnknownMessage)
	if !ok {
		t.Fatalf("got %T, want *UnknownMessage", msg)
	}
	if um.RawType != "message/system" {
		t.Errorf("RawType = %q, want %q", um.RawType, "message/system")
	}
}

// TestParseRealGeminiOutput verifies parsing against actual gemini CLI output.
func TestParseRealGeminiOutput(t *testing.T) {
	lines := []struct {
		json     string
		wantType string
	}{
		{`{"type":"init","timestamp":"2026-03-24T09:23:50.200Z","session_id":"5f2f1add-65c8-4ff0-97a2-c5251e064ada","model":"auto-gemini-3"}`, "init"},
		{`{"type":"message","timestamp":"2026-03-24T09:23:50.200Z","role":"user","content":"what files are in /tmp? just list the first 3"}`, "user"},
		{`{"type":"message","timestamp":"2026-03-24T09:23:53.972Z","role":"assistant","content":"I will list the first three items.\n","delta":true}`, "assistant"},
		{`{"type":"tool_use","timestamp":"2026-03-24T09:23:54.010Z","tool_name":"list_directory","tool_id":"list_directory_1774344234010_0","parameters":{"dir_path":"/tmp"}}`, "tool_use"},
		{`{"type":"tool_result","timestamp":"2026-03-24T09:23:54.399Z","tool_id":"list_directory_1774344234010_0","status":"error","output":"Path not in workspace","error":{"type":"invalid_tool_params","message":"Path not in workspace"}}`, "tool_result"},
		{`{"type":"result","timestamp":"2026-03-24T09:24:15.521Z","status":"success","stats":{"total_tokens":26930,"input_tokens":26512,"output_tokens":156,"cached":6329,"input":20183,"duration_ms":25322,"tool_calls":2,"models":{"gemini-2.5-flash-lite":{"total_tokens":2622,"input_tokens":2450,"output_tokens":48,"cached":0,"input":2450},"gemini-3-flash-preview":{"total_tokens":24308,"input_tokens":24062,"output_tokens":108,"cached":6329,"input":17733}}}}`, "result"},
	}

	for _, tt := range lines {
		msg, err := ParseMessage([]byte(tt.json))
		if err != nil {
			t.Errorf("ParseMessage(%q): %v", tt.wantType, err)
			continue
		}
		if msg.Type() != tt.wantType {
			t.Errorf("type = %q, want %q", msg.Type(), tt.wantType)
		}
	}

	// Verify specific fields from the real result event.
	resultJSON := `{"type":"result","status":"success","stats":{"total_tokens":26930,"input_tokens":26512,"output_tokens":156,"cached":6329,"input":20183,"duration_ms":25322,"tool_calls":2,"models":{"gemini-2.5-flash-lite":{"total_tokens":2622},"gemini-3-flash-preview":{"total_tokens":24308}}}}`
	msg, _ := ParseMessage([]byte(resultJSON))
	rm := msg.(*ResultMessage)
	if rm.Stats.TotalTokens != 26930 {
		t.Errorf("TotalTokens = %d, want 26930", rm.Stats.TotalTokens)
	}
	if rm.Stats.Input != 20183 {
		t.Errorf("Input = %d, want 20183", rm.Stats.Input)
	}
	if rm.Stats.Cached != 6329 {
		t.Errorf("Cached = %d, want 6329", rm.Stats.Cached)
	}
	if len(rm.Stats.Models) != 2 {
		t.Errorf("Models count = %d, want 2", len(rm.Stats.Models))
	}
	if rm.Stats.Models["gemini-2.5-flash-lite"].TotalTokens != 2622 {
		t.Errorf("flash-lite tokens = %d, want 2622", rm.Stats.Models["gemini-2.5-flash-lite"].TotalTokens)
	}
}

func TestParseErrorMessage(t *testing.T) {
	data := []byte(`{"type":"error","message":"API key invalid","timestamp":"2026-01-01T00:00:00Z"}`)
	msg, err := ParseMessage(data)
	if err != nil {
		t.Fatalf("ParseMessage: %v", err)
	}
	em, ok := msg.(*ErrorMessage)
	if !ok {
		t.Fatalf("got %T, want *ErrorMessage", msg)
	}
	if em.ErrorText != "API key invalid" {
		t.Errorf("ErrorText = %q, want %q", em.ErrorText, "API key invalid")
	}
}

func TestParseInvalidJSON(t *testing.T) {
	_, err := ParseMessage([]byte(`not json`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if _, ok := err.(*ProtocolError); !ok {
		t.Errorf("got %T, want *ProtocolError", err)
	}
}

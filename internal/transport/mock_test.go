package transport

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestMockTransportReplay(t *testing.T) {
	m := &MockTransport{
		RawLines: []json.RawMessage{
			json.RawMessage(`{"type":"init"}`),
			json.RawMessage(`{"type":"result"}`),
		},
	}
	if err := m.Start(context.Background(), "test", nil); err != nil {
		t.Fatalf("Start: %v", err)
	}

	var count int
	for msg := range m.Lines() {
		if msg.Err != nil {
			t.Fatalf("unexpected error: %v", msg.Err)
		}
		count++
	}
	if count != 2 {
		t.Errorf("received %d messages, want 2", count)
	}
}

func TestMockTransportStartError(t *testing.T) {
	m := &MockTransport{StartErr: errors.New("fail")}
	if err := m.Start(context.Background(), "test", nil); err == nil {
		t.Error("expected Start error")
	}
}

func TestMockTransportEmpty(t *testing.T) {
	m := &MockTransport{}
	if err := m.Start(context.Background(), "test", nil); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for msg := range m.Lines() {
		t.Errorf("unexpected message: %v", msg)
	}
}

func TestMockTransportClose(t *testing.T) {
	m := &MockTransport{}
	if err := m.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}

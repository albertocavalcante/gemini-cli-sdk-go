// Package transport abstracts CLI subprocess communication.
package transport

import (
	"context"
	"encoding/json"
)

// Options holds the configuration passed from the public API to the transport layer.
type Options struct {
	Model                 string
	ApprovalMode          string
	Sandbox               bool
	SessionID             string
	Extensions            []string
	AllowedMCPServerNames []string
	IncludeDirectories    []string
	Policy                []string
	WorkingDirectory      string
	CLIPath               string
	Env                   map[string]string
	SettingsPath          string
}

// RawLineOrError carries either a raw JSON line or an error.
type RawLineOrError struct {
	Line []byte
	Err  error
}

// Transport abstracts the mechanism for communicating with the Gemini CLI.
type Transport interface {
	// Start launches the CLI process with the given prompt and options.
	Start(ctx context.Context, prompt string, opts *Options) error

	// Lines returns a channel that delivers raw JSON lines from the CLI's stdout.
	// The channel is closed when the process exits or an error occurs.
	Lines() <-chan RawLineOrError

	// Close terminates the CLI process and cleans up resources.
	Close() error
}

// MockTransport replays canned JSON lines for testing.
type MockTransport struct {
	RawLines []json.RawMessage
	StartErr error
	ch       chan RawLineOrError
}

// Start replays the canned lines.
func (m *MockTransport) Start(_ context.Context, _ string, _ *Options) error {
	if m.StartErr != nil {
		return m.StartErr
	}
	m.ch = make(chan RawLineOrError, len(m.RawLines)+1)
	for _, line := range m.RawLines {
		cp := make([]byte, len(line))
		copy(cp, line)
		m.ch <- RawLineOrError{Line: cp}
	}
	close(m.ch)
	return nil
}

// Lines returns the replay channel.
func (m *MockTransport) Lines() <-chan RawLineOrError {
	return m.ch
}

// Close is a no-op for the mock.
func (m *MockTransport) Close() error { return nil }

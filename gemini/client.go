package gemini

import (
	"context"
	"fmt"
	"sync"

	"github.com/albertocavalcante/gemini-cli-sdk-go/internal/transport"
)

// GeminiClient manages a persistent Gemini CLI session.
// Session state is preserved across [GeminiClient.Query] calls via --resume.
type GeminiClient struct {
	opts      Options
	sessionID string
	mu        sync.Mutex
	// transport is nil for production (creates SubprocessTransport per query).
	// Set via newClientWithTransport for testing.
	transport transport.Transport
}

// NewClient creates a new GeminiClient with the given options.
func NewClient(opts Options) *GeminiClient {
	return &GeminiClient{opts: opts}
}

// SessionID returns the current session ID (thread-safe).
func (c *GeminiClient) SessionID() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.sessionID
}

// Query sends a prompt and returns a channel of messages.
// Subsequent calls use --resume with the captured session ID.
func (c *GeminiClient) Query(ctx context.Context, prompt string) <-chan MessageOrError {
	c.mu.Lock()
	opts := c.opts
	if c.sessionID != "" {
		opts.SessionID = c.sessionID
	}
	c.mu.Unlock()

	tOpts, cleanup, err := toTransportOptions(&opts)
	if err != nil {
		ch := make(chan MessageOrError, 1)
		go func() {
			ch <- MessageOrError{Err: fmt.Errorf("configuring transport: %w", err)}
			close(ch)
		}()
		return ch
	}

	tr := c.newTransport()
	ch := make(chan MessageOrError, 10)
	hr := newHookRunner(opts.Hooks, c.SessionID())

	go func() {
		defer close(ch)
		defer cleanup()
		defer func() { _ = tr.Close() }()

		if err := tr.Start(ctx, prompt, tOpts); err != nil {
			ch <- MessageOrError{Err: err}
			return
		}

		for raw := range tr.Lines() {
			if ctx.Err() != nil {
				return
			}
			if raw.Err != nil {
				ch <- MessageOrError{Err: raw.Err}
				continue
			}
			msg, err := ParseMessage(raw.Line)
			if err != nil {
				ch <- MessageOrError{Err: err}
				continue
			}

			// Capture session ID from init event.
			if im, ok := msg.(*InitMessage); ok && im.SessionID != "" {
				c.mu.Lock()
				c.sessionID = im.SessionID
				c.mu.Unlock()
				hr.sessionID = im.SessionID
			}

			hr.fireHooks(ctx, msg)
			ch <- MessageOrError{Message: msg}
		}
	}()

	return ch
}

// newTransport returns the transport to use for a query.
func (c *GeminiClient) newTransport() transport.Transport {
	if c.transport != nil {
		return c.transport
	}
	return &transport.SubprocessTransport{}
}

// newClientWithTransport creates a client with an injected transport (testing).
func newClientWithTransport(opts Options, t transport.Transport) *GeminiClient {
	return &GeminiClient{
		opts:      opts,
		transport: t,
	}
}

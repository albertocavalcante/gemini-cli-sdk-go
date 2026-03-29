package gemini

import (
	"context"
	"fmt"
	"sync"

	"github.com/albertocavalcante/gemini-cli-sdk-go/internal/transport"
)

// GeminiClient manages a persistent Gemini CLI session.
// Session state is preserved across [GeminiClient.Query] calls via --resume.
//
// When a new Query arrives while one is already in-flight, the client
// cancels the in-flight query and starts the new one immediately
// (cancel-and-replace). This prevents "session already in use" errors
// and provides snappy UX when users send messages rapidly.
type GeminiClient struct {
	opts      Options
	sessionID string
	mu        sync.RWMutex
	// transport is nil for production (creates SubprocessTransport per query).
	// Set via newClientWithTransport for testing.
	transport transport.Transport

	// Active query tracking for cancel-and-replace.
	activeCancel context.CancelFunc
	activeDone   chan struct{}
}

// NewClient creates a new GeminiClient with the given options.
func NewClient(opts Options) *GeminiClient {
	return &GeminiClient{opts: opts}
}

// SessionID returns the current session ID (thread-safe).
func (c *GeminiClient) SessionID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.sessionID
}

// Close cancels any in-flight query and releases resources.
func (c *GeminiClient) Close() error {
	c.mu.Lock()
	cancel := c.activeCancel
	done := c.activeDone
	c.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	if done != nil {
		<-done
	}
	return nil
}

// Query sends a prompt and returns a channel of messages.
// Subsequent calls use --resume with the captured session ID.
//
// If a previous query is still in-flight, it is cancelled before the new
// one starts. This ensures only one subprocess uses the session at a time.
func (c *GeminiClient) Query(ctx context.Context, prompt string) <-chan MessageOrError {
	ch := make(chan MessageOrError, 10)

	go func() {
		defer close(ch)

		// Cancel any in-flight query and wait for it to finish.
		c.mu.Lock()
		prevCancel := c.activeCancel
		prevDone := c.activeDone
		c.mu.Unlock()

		if prevCancel != nil {
			prevCancel()
		}
		if prevDone != nil {
			<-prevDone
		}

		// Create a cancellable context for this query.
		queryCtx, queryCancel := context.WithCancel(ctx)
		done := make(chan struct{})

		// Register as the active query and snapshot session state.
		c.mu.Lock()
		c.activeCancel = queryCancel
		c.activeDone = done
		opts := c.opts
		if c.sessionID != "" {
			opts.SessionID = c.sessionID
		}
		c.mu.Unlock()

		// Signal completion when this goroutine exits.
		defer func() {
			queryCancel()
			close(done)
		}()

		tOpts, cleanup, err := toTransportOptions(&opts)
		if err != nil {
			ch <- MessageOrError{Err: fmt.Errorf("configuring transport: %w", err)}
			return
		}
		defer cleanup()

		tr := c.newTransport()
		defer func() { _ = tr.Close() }()

		if err := tr.Start(queryCtx, prompt, tOpts); err != nil {
			ch <- MessageOrError{Err: err}
			return
		}

		hr := newHookRunner(opts.Hooks, c.SessionID())

		for raw := range tr.Lines() {
			if queryCtx.Err() != nil {
				return
			}
			if raw.Err != nil {
				select {
				case ch <- MessageOrError{Err: raw.Err}:
				case <-queryCtx.Done():
					return
				}
				continue
			}
			msg, err := ParseMessage(raw.Line)
			if err != nil {
				select {
				case ch <- MessageOrError{Err: err}:
				case <-queryCtx.Done():
					return
				}
				continue
			}

			// Capture session ID from init event.
			if im, ok := msg.(*InitMessage); ok && im.SessionID != "" {
				c.mu.Lock()
				c.sessionID = im.SessionID
				c.mu.Unlock()
				hr.sessionID = im.SessionID
			}

			hr.fireHooks(queryCtx, msg)

			select {
			case ch <- MessageOrError{Message: msg}:
			case <-queryCtx.Done():
				return
			}
		}
	}()

	return ch
}

// newTransport returns the transport to use for a query.
func (c *GeminiClient) newTransport() transport.Transport {
	c.mu.RLock()
	defer c.mu.RUnlock()
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

package gemini

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/albertocavalcante/gemini-cli-sdk-go/internal/transport"
)

type cancellationTransport struct {
	lines     chan transport.RawLineOrError
	closed    chan struct{}
	delivered chan struct{}
	once      sync.Once
	raw       transport.RawLineOrError
	idle      bool
}

func (r *cancellationTransport) Start(ctx context.Context, _ string, _ *transport.Options) error {
	go func() {
		if !r.idle {
			for i := 0; i < 11; i++ {
				select {
				case r.lines <- r.raw:
				case <-ctx.Done():
					close(r.lines)
					return
				}
			}
		}
		close(r.delivered)
		<-ctx.Done()
		close(r.lines)
	}()
	return nil
}
func (r *cancellationTransport) Lines() <-chan transport.RawLineOrError { return r.lines }
func (r *cancellationTransport) Close() error                           { r.once.Do(func() { close(r.closed) }); return nil }

type signallingContext struct {
	context.Context
	calls         atomic.Int32
	eleventhCheck chan struct{}
}

func (r *signallingContext) Err() error {
	err := r.Context.Err()
	if r.calls.Add(1) == 11 {
		close(r.eleventhCheck)
	}
	return err
}
func testOneShotCancellationWithFullOutputBuffer(t *testing.T, raw transport.RawLineOrError) {
	base, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx := &signallingContext{Context: base, eleventhCheck: make(chan struct{})}
	mock := &cancellationTransport{lines: make(chan transport.RawLineOrError), closed: make(chan struct{}), delivered: make(chan struct{}), raw: raw}
	ch := queryWithTransport(ctx, "", Options{}, mock)
	select {
	case <-mock.delivered:
	case <-time.After(time.Second):
		t.Fatal("fake transport did not deliver 11 messages")
	}
	select {
	case <-ctx.eleventhCheck:
	case <-time.After(time.Second):
		t.Fatal("query did not check context for 11th message")
	}
	cancel()
	select {
	case <-mock.closed:
	case <-time.After(500 * time.Millisecond):
		t.Error("cancelled Query blocked sending its 11th public message; transport Close was never called")
	}
	for range ch {
	}
}

func TestOneShotCancellationWithFullOutputBuffer(t *testing.T) {
	for _, test := range []struct {
		name string
		raw  transport.RawLineOrError
	}{
		{"message", transport.RawLineOrError{Line: []byte(`{"type":"message","role":"assistant","content":"hi"}`)}},
		{"protocol-error", transport.RawLineOrError{Line: []byte(`invalid JSON`)}},
		{"transport-error", transport.RawLineOrError{Err: errors.New("read failed")}},
	} {
		t.Run(test.name, func(t *testing.T) { testOneShotCancellationWithFullOutputBuffer(t, test.raw) })
	}
}
func TestOneShotCancellationWhileAwaitingOutput(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mock := &cancellationTransport{lines: make(chan transport.RawLineOrError), closed: make(chan struct{}), delivered: make(chan struct{}), idle: true}
	ch := queryWithTransport(ctx, "", Options{}, mock)
	select {
	case <-mock.delivered:
	case <-time.After(time.Second):
		t.Fatal("transport not started")
	}
	cancel()
	select {
	case <-mock.closed:
	case <-time.After(time.Second):
		t.Fatal("idle query did not release transport")
	}
	for range ch {
	}
}

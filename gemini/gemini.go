// Package gemini provides a Go SDK for the Gemini CLI.
//
// It spawns the gemini CLI as a subprocess and streams structured messages
// back over a channel, providing a Go-idiomatic interface for building
// agents and tools powered by Gemini.
//
// Basic usage:
//
//	ctx := context.Background()
//	for msg := range gemini.Query(ctx, "Hello, Gemini!", gemini.Options{}) {
//	    if msg.Err != nil {
//	        log.Fatal(msg.Err)
//	    }
//	    fmt.Println(msg.Message.Type())
//	}
package gemini

import (
	"context"
	"fmt"

	"github.com/albertocavalcante/gemini-cli-sdk-go/internal/transport"
)

// MessageOrError carries either a parsed Message or an error.
// Callers should check Err first.
type MessageOrError struct {
	Message Message
	Err     error
}

// Query sends a one-shot prompt to the Gemini CLI and returns a channel
// of streamed messages. The channel is closed when the CLI process exits.
//
// Each invocation spawns a new subprocess. For multi-turn sessions,
// use [NewClient] instead.
func Query(ctx context.Context, prompt string, opts Options) <-chan MessageOrError {
	return queryWithTransport(ctx, prompt, opts, &transport.SubprocessTransport{})
}

// queryWithTransport is the internal implementation that accepts a custom
// transport (used for testing with MockTransport).
func queryWithTransport(ctx context.Context, prompt string, opts Options, t transport.Transport) <-chan MessageOrError {
	ch := make(chan MessageOrError, 10)
	send := func(item MessageOrError) bool {
		select {
		case ch <- item:
			return true
		case <-ctx.Done():
			return false
		}
	}

	tOpts, cleanup, err := toTransportOptions(&opts)
	if err != nil {
		go func() {
			send(MessageOrError{Err: fmt.Errorf("configuring transport: %w", err)})
			close(ch)
		}()
		return ch
	}

	go func() {
		defer close(ch)
		defer cleanup()
		defer func() { _ = t.Close() }()

		if err := t.Start(ctx, prompt, tOpts); err != nil {
			send(MessageOrError{Err: err})
			return
		}

		lines := t.Lines()
		for {
			var raw transport.RawLineOrError
			select {
			case item, ok := <-lines:
				if !ok {
					return
				}
				raw = item
			case <-ctx.Done():
				return
			}
			if ctx.Err() != nil {
				return
			}
			if raw.Err != nil {
				if !send(MessageOrError{Err: raw.Err}) {
					return
				}
				continue
			}
			msg, err := ParseMessage(raw.Line)
			if err != nil {
				if !send(MessageOrError{Err: err}) {
					return
				}
				continue
			}
			if !send(MessageOrError{Message: msg}) {
				return
			}
		}
	}()

	return ch
}

// toTransportOptions converts public Options to internal transport Options.
// Returns the options, a cleanup function, and any configuration error.
func toTransportOptions(opts *Options) (*transport.Options, func(), error) {
	tOpts := &transport.Options{
		Model:                 opts.Model,
		ApprovalMode:          opts.ApprovalMode,
		Sandbox:               opts.Sandbox,
		SessionID:             opts.SessionID,
		Extensions:            opts.Extensions,
		AllowedMCPServerNames: opts.AllowedMCPServerNames,
		IncludeDirectories:    opts.IncludeDirectories,
		Policy:                opts.Policy,
		WorkingDirectory:      opts.WorkingDirectory,
		CLIPath:               opts.CLIPath,
		CLIPrefixArgs:         opts.CLIPrefixArgs,
		Env:                   opts.Env,
	}

	noop := func() {}

	switch {
	case opts.MCPSettingsPath != "":
		tOpts.SettingsPath = opts.MCPSettingsPath
		return tOpts, noop, nil
	case len(opts.MCPServers) > 0:
		dir, err := WriteSettingsDir(opts.MCPServers)
		if err != nil {
			return nil, noop, fmt.Errorf("writing MCP settings: %w", err)
		}
		tOpts.SettingsPath = dir
		return tOpts, func() { _ = CleanupSettingsDir(dir) }, nil
	default:
		return tOpts, noop, nil
	}
}

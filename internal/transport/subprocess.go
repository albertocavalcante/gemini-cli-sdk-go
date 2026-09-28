package transport

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const (
	defaultCLI  = "gemini"
	maxLineSize = 1 << 20 // 1 MB
	chanBufSize = 64
)

// SubprocessTransport spawns the gemini CLI as a subprocess and
// reads streaming JSON from its stdout.
type SubprocessTransport struct {
	cmd       *exec.Cmd
	ch        chan RawLineOrError
	done      chan struct{}
	stop      chan struct{}
	stdout    io.ReadCloser
	stderr    *strings.Builder
	closeOnce sync.Once
	mu        sync.Mutex
}

// Start launches the gemini CLI with the given prompt and options.
func (s *SubprocessTransport) Start(ctx context.Context, prompt string, opts *Options) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.done != nil {
		return fmt.Errorf("transport already started")
	}
	cliPath := defaultCLI
	if opts != nil && opts.CLIPath != "" {
		cliPath = opts.CLIPath
	}
	resolved, err := LookPath(cliPath)
	if err != nil {
		return fmt.Errorf("cannot find %s: %w", cliPath, err)
	}
	cmd := exec.CommandContext(ctx, resolved, buildArgs(prompt, opts)...)
	s.cmd = cmd
	if opts != nil && opts.WorkingDirectory != "" {
		cmd.Dir = opts.WorkingDirectory
	}
	cmd.Env = buildEnv(opts)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}
	s.stdout = stdout
	s.stderr = &strings.Builder{}
	cmd.Stderr = s.stderr
	// Cancellation must also unblock a scanner when descendants inherit stdout.
	cmd.Cancel = func() error {
		err := cmd.Process.Kill()
		_ = stdout.Close()
		return err
	}
	// Bound stderr copier cleanup when descendants inherit that descriptor.
	cmd.WaitDelay = time.Second
	if err := cmd.Start(); err != nil {
		_ = stdout.Close()
		return fmt.Errorf("starting %s: %w", cliPath, err)
	}
	s.ch = make(chan RawLineOrError, chanBufSize)
	s.done = make(chan struct{})
	s.stop = make(chan struct{})
	go func() {
		defer close(s.done)
		defer close(s.ch)
		send := func(raw RawLineOrError) bool {
			select {
			case <-ctx.Done():
				return false
			case <-s.stop:
				return false
			default:
			}
			select {
			case s.ch <- raw:
				return true
			case <-ctx.Done():
				return false
			case <-s.stop:
				return false
			}
		}
		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 0, 64*1024), maxLineSize)
		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}
			if !send(RawLineOrError{Line: bytes.Clone(line)}) {
				break
			}
		}
		scanErr := scanner.Err()
		_ = stdout.Close()
		if scanErr != nil {
			_ = cmd.Process.Kill()
		}
		// Wait always reaps the process before attempting terminal error delivery.
		waitErr := cmd.Wait()
		if scanErr != nil {
			send(RawLineOrError{Err: fmt.Errorf("scanner: %w", scanErr)})
		}
		if waitErr != nil {
			send(RawLineOrError{Err: fmt.Errorf("process exited: %w (stderr: %s)", waitErr, strings.TrimSpace(s.stderr.String()))})
		}
	}()
	return nil
}

// Lines returns the channel delivering raw JSON lines.
func (s *SubprocessTransport) Lines() <-chan RawLineOrError {
	return s.ch
}

// Close terminates the subprocess and unblocks a reader whose output
// channel is full, even when the parent context remains active.
func (s *SubprocessTransport) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.done == nil {
		return nil
	}
	s.closeOnce.Do(func() { close(s.stop) })
	_ = s.cmd.Process.Kill()
	_ = s.stdout.Close()
	<-s.done
	return nil
}

// buildArgs constructs the CLI argument list from options.
// When CLIPrefixArgs is set (e.g., for npx), they are prepended before the
// gemini-cli flags: npx --yes @google/gemini-cli -p prompt --output-format stream-json
func buildArgs(prompt string, opts *Options) []string {
	var prefix []string
	if opts != nil {
		prefix = opts.CLIPrefixArgs
	}
	args := append(prefix, "-p", prompt, "--output-format", "stream-json")
	if opts == nil {
		return args
	}

	if opts.Model != "" {
		args = append(args, "--model", opts.Model)
	}
	if opts.ApprovalMode != "" {
		args = append(args, "--approval-mode", opts.ApprovalMode)
	}
	if opts.Sandbox {
		args = append(args, "--sandbox")
	}
	if opts.SessionID != "" {
		args = append(args, "--resume", opts.SessionID)
	}
	for _, ext := range opts.Extensions {
		args = append(args, "--extensions", ext)
	}
	if len(opts.AllowedMCPServerNames) > 0 {
		args = append(args, "--allowed-mcp-server-names", strings.Join(opts.AllowedMCPServerNames, ","))
	}
	if len(opts.IncludeDirectories) > 0 {
		args = append(args, "--include-directories", strings.Join(opts.IncludeDirectories, ","))
	}
	for _, p := range opts.Policy {
		args = append(args, "--policy", p)
	}

	return args
}

// buildEnv constructs the environment variable list for the subprocess.
// Returns nil (inherit parent env) when no custom vars are needed.
func buildEnv(opts *Options) []string {
	if opts == nil {
		return nil
	}

	needsCustomEnv := len(opts.Env) > 0 || opts.SettingsPath != ""
	if !needsCustomEnv {
		return nil
	}

	env := os.Environ()
	for k, v := range opts.Env {
		env = append(env, k+"="+v)
	}
	if opts.SettingsPath != "" {
		env = append(env, "GEMINI_HOME="+opts.SettingsPath)
	}
	return env
}

package transport

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	defaultCLI  = "gemini"
	maxLineSize = 1 << 20 // 1 MB
	chanBufSize = 64
)

// SubprocessTransport spawns the gemini CLI as a subprocess and
// reads streaming JSON from its stdout.
type SubprocessTransport struct {
	cmd    *exec.Cmd
	ch     chan RawLineOrError
	done   chan struct{}
	stderr *strings.Builder
}

// Start launches the gemini CLI with the given prompt and options.
func (s *SubprocessTransport) Start(ctx context.Context, prompt string, opts *Options) error {
	cliPath := defaultCLI
	if opts != nil && opts.CLIPath != "" {
		cliPath = opts.CLIPath
	}

	s.cmd = exec.CommandContext(ctx, cliPath, buildArgs(prompt, opts)...)

	if opts != nil && opts.WorkingDirectory != "" {
		s.cmd.Dir = opts.WorkingDirectory
	}

	// Build environment: inherit current + custom vars + settings path.
	s.cmd.Env = buildEnv(opts)

	stdout, err := s.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}

	s.stderr = &strings.Builder{}
	s.cmd.Stderr = s.stderr

	if err := s.cmd.Start(); err != nil {
		return fmt.Errorf("starting %s: %w", cliPath, err)
	}

	s.ch = make(chan RawLineOrError, chanBufSize)
	s.done = make(chan struct{})

	go func() {
		defer close(s.done)
		defer close(s.ch)

		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 0, 64*1024), maxLineSize)

		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}
			s.ch <- RawLineOrError{Line: bytes.Clone(line)}
		}

		if err := scanner.Err(); err != nil {
			s.ch <- RawLineOrError{Err: fmt.Errorf("scanner: %w", err)}
		}

		if err := s.cmd.Wait(); err != nil {
			s.ch <- RawLineOrError{Err: fmt.Errorf("process exited: %w (stderr: %s)", err, strings.TrimSpace(s.stderr.String()))}
		}
	}()

	return nil
}

// Lines returns the channel delivering raw JSON lines.
func (s *SubprocessTransport) Lines() <-chan RawLineOrError {
	return s.ch
}

// Close terminates the subprocess if still running.
func (s *SubprocessTransport) Close() error {
	if s.cmd == nil || s.cmd.Process == nil {
		return nil
	}
	_ = s.cmd.Process.Kill()
	if s.done != nil {
		<-s.done
	}
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

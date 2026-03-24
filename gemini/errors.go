package gemini

import (
	"errors"
	"fmt"
)

// CLIError is returned when the Gemini CLI returns a non-zero exit code.
type CLIError struct {
	Message  string
	Stderr   string
	ExitCode int
}

func (e *CLIError) Error() string {
	return fmt.Sprintf("gemini cli error (exit %d): %s", e.ExitCode, e.Message)
}

// ProtocolError is returned when the SDK cannot parse the CLI's JSON output.
type ProtocolError struct {
	Message string
	Raw     []byte
}

func (e *ProtocolError) Error() string {
	return fmt.Sprintf("protocol error: %s", e.Message)
}

// ProcessError is returned when the CLI subprocess exits abnormally.
type ProcessError struct {
	Message  string
	ExitCode int
}

func (e *ProcessError) Error() string {
	return fmt.Sprintf("process error (exit %d): %s", e.ExitCode, e.Message)
}

// IsInvalidInput reports whether err indicates invalid input (exit code 42).
func IsInvalidInput(err error) bool {
	return hasExitCode(err, ExitInvalidInput)
}

// IsTurnLimitExceeded reports whether err indicates turn limit exceeded (exit code 53).
func IsTurnLimitExceeded(err error) bool {
	return hasExitCode(err, ExitTurnLimit)
}

// hasExitCode checks if an error chain contains a CLIError or ProcessError
// with the given exit code.
func hasExitCode(err error, code int) bool {
	var cliErr *CLIError
	if errors.As(err, &cliErr) && cliErr.ExitCode == code {
		return true
	}
	var procErr *ProcessError
	return errors.As(err, &procErr) && procErr.ExitCode == code
}

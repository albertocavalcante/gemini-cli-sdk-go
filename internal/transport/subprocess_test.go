package transport

import (
	"testing"
)

func TestBuildArgsMinimal(t *testing.T) {
	args := buildArgs("hello", nil)
	want := []string{"-p", "hello", "--output-format", "stream-json"}
	assertArgs(t, args, want)
}

func TestBuildArgsWithModel(t *testing.T) {
	args := buildArgs("test", &Options{Model: "flash"})
	assertContains(t, args, "--model", "flash")
}

func TestBuildArgsWithApprovalMode(t *testing.T) {
	args := buildArgs("test", &Options{ApprovalMode: "yolo"})
	assertContains(t, args, "--approval-mode", "yolo")
}

func TestBuildArgsWithSandbox(t *testing.T) {
	args := buildArgs("test", &Options{Sandbox: true})
	assertHas(t, args, "--sandbox")
}

func TestBuildArgsWithResume(t *testing.T) {
	args := buildArgs("test", &Options{SessionID: "sess_123"})
	assertContains(t, args, "--resume", "sess_123")
}

func TestBuildArgsWithExtensions(t *testing.T) {
	args := buildArgs("test", &Options{Extensions: []string{"ext1", "ext2"}})
	assertContains(t, args, "--extensions", "ext1")
	assertContains(t, args, "--extensions", "ext2")
}

func TestBuildArgsWithAllowedMCPServers(t *testing.T) {
	args := buildArgs("test", &Options{AllowedMCPServerNames: []string{"srv1", "srv2"}})
	assertContains(t, args, "--allowed-mcp-server-names", "srv1,srv2")
}

func TestBuildArgsWithIncludeDirectories(t *testing.T) {
	args := buildArgs("test", &Options{IncludeDirectories: []string{"/dir1", "/dir2"}})
	assertContains(t, args, "--include-directories", "/dir1,/dir2")
}

func TestBuildArgsWithPolicy(t *testing.T) {
	args := buildArgs("test", &Options{Policy: []string{"p1", "p2"}})
	assertContains(t, args, "--policy", "p1")
	assertContains(t, args, "--policy", "p2")
}

func TestBuildArgsPromptPosition(t *testing.T) {
	args := buildArgs("my prompt", &Options{Model: "pro"})
	// -p and prompt must be args[0] and args[1]
	if args[0] != "-p" || args[1] != "my prompt" {
		t.Errorf("prompt must be first: got args[0]=%q args[1]=%q", args[0], args[1])
	}
}

func TestBuildArgsFullOptions(t *testing.T) {
	args := buildArgs("full test", &Options{
		Model:                 "gemini-2.5-pro",
		ApprovalMode:          "auto_edit",
		Sandbox:               true,
		SessionID:             "latest",
		Extensions:            []string{"ext1"},
		AllowedMCPServerNames: []string{"srv"},
		IncludeDirectories:    []string{"/tmp"},
		Policy:                []string{"strict"},
	})

	assertContains(t, args, "--model", "gemini-2.5-pro")
	assertContains(t, args, "--approval-mode", "auto_edit")
	assertHas(t, args, "--sandbox")
	assertContains(t, args, "--resume", "latest")
	assertContains(t, args, "--extensions", "ext1")
	assertContains(t, args, "--allowed-mcp-server-names", "srv")
	assertContains(t, args, "--include-directories", "/tmp")
	assertContains(t, args, "--policy", "strict")
}

func TestSubprocessCloseWithoutStart(t *testing.T) {
	s := &SubprocessTransport{}
	if err := s.Close(); err != nil {
		t.Errorf("Close without Start should not error: %v", err)
	}
}

func TestSubprocessLinesNil(t *testing.T) {
	s := &SubprocessTransport{}
	if s.Lines() != nil {
		t.Error("Lines before Start should be nil")
	}
}

// --- helpers ---

func assertArgs(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("args length = %d, want %d\ngot:  %v\nwant: %v", len(got), len(want), got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("args[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func assertContains(t *testing.T, args []string, flag, value string) {
	t.Helper()
	for i, a := range args {
		if a == flag && i+1 < len(args) && args[i+1] == value {
			return
		}
	}
	t.Errorf("args missing %s %s: %v", flag, value, args)
}

func assertHas(t *testing.T, args []string, flag string) {
	t.Helper()
	for _, a := range args {
		if a == flag {
			return
		}
	}
	t.Errorf("args missing %s: %v", flag, args)
}

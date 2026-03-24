package gemini

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteSettingsDir(t *testing.T) {
	dir, err := WriteSettingsDir([]MCPServerConfig{
		{Name: "my-server", Command: "node", Args: []string{"server.js"}, Env: map[string]string{"PORT": "3000"}},
	})
	if err != nil {
		t.Fatalf("WriteSettingsDir: %v", err)
	}
	defer CleanupSettingsDir(dir)

	data, err := os.ReadFile(filepath.Join(dir, "settings.json"))
	if err != nil {
		t.Fatalf("reading settings.json: %v", err)
	}

	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("parsing settings.json: %v", err)
	}

	servers, ok := settings["mcpServers"].(map[string]any)
	if !ok {
		t.Fatal("mcpServers not found in settings.json")
	}
	if _, ok := servers["my-server"]; !ok {
		t.Error("my-server not found in mcpServers")
	}
}

func TestWriteSettingsDirMultiple(t *testing.T) {
	dir, err := WriteSettingsDir([]MCPServerConfig{
		{Name: "srv1", Command: "cmd1"},
		{Name: "srv2", Command: "cmd2"},
	})
	if err != nil {
		t.Fatalf("WriteSettingsDir: %v", err)
	}
	defer CleanupSettingsDir(dir)

	data, _ := os.ReadFile(filepath.Join(dir, "settings.json"))
	var settings mcpSettingsJSON
	json.Unmarshal(data, &settings)

	if len(settings.MCPServers) != 2 {
		t.Errorf("got %d servers, want 2", len(settings.MCPServers))
	}
}

func TestWriteSettingsDirEmpty(t *testing.T) {
	dir, err := WriteSettingsDir(nil)
	if err != nil {
		t.Fatalf("WriteSettingsDir: %v", err)
	}
	if dir != "" {
		t.Errorf("dir = %q, want empty for nil servers", dir)
	}
}

func TestWriteSettingsDirMissingName(t *testing.T) {
	_, err := WriteSettingsDir([]MCPServerConfig{
		{Command: "cmd"},
	})
	if err == nil {
		t.Fatal("expected error for missing Name")
	}
}

func TestWriteSettingsDirPermissions(t *testing.T) {
	dir, err := WriteSettingsDir([]MCPServerConfig{
		{Name: "test", Command: "echo"},
	})
	if err != nil {
		t.Fatalf("WriteSettingsDir: %v", err)
	}
	defer CleanupSettingsDir(dir)

	info, err := os.Stat(filepath.Join(dir, "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	perm := info.Mode().Perm()
	if perm != 0o600 {
		t.Errorf("permissions = %o, want 0600", perm)
	}
}

func TestCleanupSettingsDirEmpty(t *testing.T) {
	if err := CleanupSettingsDir(""); err != nil {
		t.Errorf("CleanupSettingsDir empty: %v", err)
	}
}

func TestCleanupSettingsDirNonexistent(t *testing.T) {
	if err := CleanupSettingsDir("/nonexistent/path/xyz"); err != nil {
		t.Errorf("CleanupSettingsDir nonexistent: %v", err)
	}
}

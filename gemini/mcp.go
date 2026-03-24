package gemini

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// MCPServerConfig defines an MCP server for the Gemini CLI.
type MCPServerConfig struct {
	Name    string
	Command string
	Args    []string
	Env     map[string]string
	CWD     string
}

// mcpSettingsJSON matches the gemini CLI's settings.json schema for MCP servers.
type mcpSettingsJSON struct {
	MCPServers map[string]mcpServerJSON `json:"mcpServers"`
}

type mcpServerJSON struct {
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	CWD     string            `json:"cwd,omitempty"`
}

// WriteSettingsDir creates a temporary directory with a settings.json
// containing the given MCP server configurations. Returns the directory
// path (suitable for GEMINI_HOME env var).
func WriteSettingsDir(servers []MCPServerConfig) (string, error) {
	if len(servers) == 0 {
		return "", nil
	}

	settings := mcpSettingsJSON{
		MCPServers: make(map[string]mcpServerJSON, len(servers)),
	}
	for _, s := range servers {
		name := s.Name
		if name == "" {
			return "", fmt.Errorf("MCP server config missing Name")
		}
		settings.MCPServers[name] = mcpServerJSON{
			Command: s.Command,
			Args:    s.Args,
			Env:     s.Env,
			CWD:     s.CWD,
		}
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling settings: %w", err)
	}

	dir, err := os.MkdirTemp("", fmt.Sprintf("gemini-sdk-%d-*", os.Getpid()))
	if err != nil {
		return "", fmt.Errorf("creating temp dir: %w", err)
	}

	settingsPath := filepath.Join(dir, "settings.json")
	if err := os.WriteFile(settingsPath, data, 0o600); err != nil {
		_ = os.RemoveAll(dir)
		return "", fmt.Errorf("writing settings.json: %w", err)
	}

	return dir, nil
}

// CleanupSettingsDir removes a temporary settings directory.
// Safe to call with an empty path.
func CleanupSettingsDir(dir string) error {
	if dir == "" {
		return nil
	}
	return os.RemoveAll(dir)
}

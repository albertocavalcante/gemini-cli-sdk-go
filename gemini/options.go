package gemini

// Options configures a Gemini CLI invocation.
type Options struct {
	// Model specifies the Gemini model to use.
	// Use Model* constants or any valid model string.
	Model string

	// ApprovalMode controls tool permission behavior.
	// Use Approval* constants: "default", "auto_edit", "yolo".
	ApprovalMode string

	// Sandbox enables sandboxed execution.
	Sandbox bool

	// WorkingDirectory sets the CWD for the CLI subprocess.
	WorkingDirectory string

	// CLIPath overrides the path to the gemini binary.
	// Defaults to "gemini" resolved from PATH.
	// When using a package runner (npx, bunx), set CLIPath to "npx"
	// and CLIPrefixArgs to ["--yes", "@google/gemini-cli"].
	CLIPath string

	// CLIPrefixArgs are inserted between CLIPath and the gemini-cli flags.
	// Used when invoking via npx/bunx/pnpm dlx.
	CLIPrefixArgs []string

	// Env provides additional environment variables for the CLI subprocess.
	// Use "GEMINI_API_KEY" for API key authentication.
	Env map[string]string

	// Hooks registers lifecycle callbacks.
	Hooks []HookRegistration

	// SessionID resumes a previous session via --resume.
	// Use "latest" for the most recent session.
	SessionID string

	// Extensions restricts which extensions are loaded.
	Extensions []string

	// AllowedMCPServerNames restricts which MCP servers can be used.
	AllowedMCPServerNames []string

	// IncludeDirectories adds extra workspace directories.
	IncludeDirectories []string

	// MCPServers configures external MCP servers.
	// Written to a temporary settings directory for the CLI.
	MCPServers []MCPServerConfig

	// MCPSettingsPath is a path to a pre-existing settings directory.
	// When set, MCPServers is ignored.
	MCPSettingsPath string

	// Policy specifies additional policy files or directories.
	Policy []string
}

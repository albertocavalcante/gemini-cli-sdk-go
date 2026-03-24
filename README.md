# gemini-cli-sdk-go

Go SDK for the [Gemini CLI](https://github.com/google-gemini/gemini-cli). Spawns `gemini` as a subprocess and streams structured JSON messages back over a Go channel.

## Install

```bash
go get github.com/albertocavalcante/gemini-cli-sdk-go
```

Requires the Gemini CLI to be installed:

```bash
npm install -g @google/gemini-cli
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/albertocavalcante/gemini-cli-sdk-go/gemini"
)

func main() {
    ctx := context.Background()
    for msg := range gemini.Query(ctx, "What is Go?", gemini.Options{}) {
        if msg.Err != nil {
            fmt.Fprintln(os.Stderr, msg.Err)
            os.Exit(1)
        }
        if am, ok := msg.Message.(*gemini.AssistantMessage); ok {
            fmt.Print(am.Content)
        }
    }
}
```

## API

### One-shot Query

```go
ch := gemini.Query(ctx, "prompt", gemini.Options{
    Model: gemini.ModelFlash,
})
for msg := range ch {
    // handle msg.Message or msg.Err
}
```

### Persistent Session (multi-turn)

```go
client := gemini.NewClient(gemini.Options{
    Model: gemini.ModelPro,
})
for msg := range client.Query(ctx, "first message") {
    // ...
}
// Session ID captured from init event, used for --resume on next call.
for msg := range client.Query(ctx, "follow-up") {
    // ...
}
```

### Lifecycle Hooks

```go
client := gemini.NewClient(gemini.Options{
    Hooks: []gemini.HookRegistration{
        {
            Event: gemini.HookPreToolUse,
            Callback: func(ctx context.Context, ev gemini.HookInput) (gemini.HookOutput, error) {
                fmt.Printf("Tool: %s\n", ev.ToolName)
                return gemini.HookOutput{}, nil
            },
        },
    },
})
```

### MCP Servers

```go
client := gemini.NewClient(gemini.Options{
    MCPServers: []gemini.MCPServerConfig{
        {
            Name:    "filesystem",
            Command: "npx",
            Args:    []string{"-y", "@modelcontextprotocol/server-filesystem", "/tmp"},
        },
    },
})
```

## Message Types

| Type | Go Type | Description |
|------|---------|-------------|
| `init` | `*InitMessage` | Session initialization (session_id, model) |
| `message` (assistant) | `*AssistantMessage` | Model response text |
| `message` (user) | `*UserMessage` | User turn |
| `tool_use` | `*ToolUseMessage` | Tool invocation (name, id, parameters) |
| `tool_result` | `*ToolResultMessage` | Tool result (status, output) |
| `result` | `*ResultMessage` | Final stats (tokens, duration, tool calls) |
| (unknown) | `*UnknownMessage` | Forward-compatible wrapper |

## Options

| Field | CLI Flag | Description |
|-------|----------|-------------|
| `Model` | `--model` | Model selection (auto/pro/flash/flash-lite) |
| `ApprovalMode` | `--approval-mode` | Tool permission mode |
| `Sandbox` | `--sandbox` | Enable sandboxed execution |
| `SessionID` | `--resume` | Resume previous session |
| `Extensions` | `--extensions` | Extension filter |
| `MCPServers` | (settings.json) | MCP server configs |
| `CLIPath` | — | Path to gemini binary |
| `Env` | — | Extra environment variables |
| `WorkingDirectory` | — | Subprocess working directory |

## Build & Test

```bash
just check    # fmt + lint + test + build
just test     # go test -race -count=1 ./...
just lint     # go vet ./...
```

## License

MIT

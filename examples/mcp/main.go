// Command mcp demonstrates MCP server integration.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/albertocavalcante/gemini-cli-sdk-go/gemini"
)

func main() {
	prompt := "List the files in /tmp using the filesystem tools"
	if len(os.Args) > 1 {
		prompt = os.Args[1]
	}

	client := gemini.NewClient(gemini.Options{
		ApprovalMode: gemini.ApprovalYolo,
		MCPServers: []gemini.MCPServerConfig{
			{
				Name:    "filesystem",
				Command: "npx",
				Args:    []string{"-y", "@modelcontextprotocol/server-filesystem", "/tmp"},
			},
		},
		Hooks: []gemini.HookRegistration{
			{
				Event: gemini.HookPreToolUse,
				Callback: func(_ context.Context, ev gemini.HookInput) (gemini.HookOutput, error) {
					fmt.Printf("[mcp] tool: %s\n", ev.ToolName)
					return gemini.HookOutput{}, nil
				},
			},
		},
	})

	ctx := context.Background()
	for msg := range client.Query(ctx, prompt) {
		if msg.Err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", msg.Err)
			os.Exit(1)
		}
		if am, ok := msg.Message.(*gemini.AssistantMessage); ok {
			fmt.Print(am.Content)
		}
	}
	fmt.Println()
}

// Command hooks demonstrates the lifecycle hook system.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/albertocavalcante/gemini-cli-sdk-go/gemini"
)

func main() {
	prompt := "List the files in the current directory"
	if len(os.Args) > 1 {
		prompt = os.Args[1]
	}

	client := gemini.NewClient(gemini.Options{
		ApprovalMode: gemini.ApprovalYolo,
		Hooks: []gemini.HookRegistration{
			{
				Event: gemini.HookPreToolUse,
				Callback: func(_ context.Context, ev gemini.HookInput) (gemini.HookOutput, error) {
					fmt.Printf("[hook] tool call: %s\n", ev.ToolName)
					return gemini.HookOutput{}, nil
				},
			},
			{
				Event: gemini.HookPostToolUse,
				Callback: func(_ context.Context, ev gemini.HookInput) (gemini.HookOutput, error) {
					fmt.Printf("[hook] tool result: %s (output: %d bytes)\n", ev.ToolName, len(ev.ToolOutput))
					return gemini.HookOutput{}, nil
				},
			},
			{
				Event: gemini.HookResult,
				Callback: func(_ context.Context, ev gemini.HookInput) (gemini.HookOutput, error) {
					if rm, ok := ev.Message.(*gemini.ResultMessage); ok {
						fmt.Printf("[hook] completed: %d tokens, %d tool calls\n",
							rm.Stats.TotalTokens, rm.Stats.ToolCalls)
					}
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

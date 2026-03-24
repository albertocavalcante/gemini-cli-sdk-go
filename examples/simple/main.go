// Command simple demonstrates a one-shot query to Gemini.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/albertocavalcante/gemini-cli-sdk-go/gemini"
)

func main() {
	prompt := "What is Go? Answer in one sentence."
	if len(os.Args) > 1 {
		prompt = os.Args[1]
	}

	ctx := context.Background()
	for msg := range gemini.Query(ctx, prompt, gemini.Options{}) {
		if msg.Err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", msg.Err)
			os.Exit(1)
		}
		switch m := msg.Message.(type) {
		case *gemini.AssistantMessage:
			fmt.Print(m.Content)
		case *gemini.ResultMessage:
			fmt.Printf("\n\n--- tokens: %d, duration: %.0fms ---\n",
				m.Stats.TotalTokens, m.Stats.DurationMs)
		}
	}
}

package execution

import (
	"context"
	"sync"
	"testing"

	anyllm "github.com/mozilla-ai/any-llm-go/providers"
)

func TestConvertMessages_ConcurrentSafe(t *testing.T) {
	msgs := []anyllm.Message{
		{Role: "system", Content: "be helpful"},
		{Role: "user", Content: "hello"},
		{
			Role:    "assistant",
			Content: "I'll help",
			ToolCalls: []anyllm.ToolCall{
				{ID: "c1", Type: "function", Function: anyllm.FunctionCall{Name: "calc", Arguments: `{"x":1}`}},
			},
		},
		{Role: "tool", ToolCallID: "c1", Content: "result"},
	}

	var wg sync.WaitGroup
	errs := make(chan string, 20)

	for range 20 {
		wg.Go(func() {
			out, system := convertMessages(msgs)
			if system != "be helpful" {
				errs <- "system wrong: " + system
				return
			}
			if len(out) != 3 {
				errs <- "messages wrong count"
			}
		})
	}

	wg.Wait()
	close(errs)

	for msg := range errs {
		t.Fatal(msg)
	}
}

func TestConfiguredProvider_ConcurrentDefaults(t *testing.T) {
	base := &simpleStubProvider{response: "ok", usage: &anyllm.Usage{PromptTokens: 5, CompletionTokens: 5}}
	p := NewConfiguredProvider(base, ProviderConfig{MaxTokens: 4096})
	ctx := context.Background()

	var wg sync.WaitGroup
	errs := make(chan error, 20)

	for range 20 {
		wg.Go(func() {
			_, err := p.Completion(ctx, anyllm.CompletionParams{Model: "test"})
			if err != nil {
				errs <- err
			}
		})
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Fatalf("concurrent completion error: %v", err)
	}
}

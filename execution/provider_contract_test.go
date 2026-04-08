package execution

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	anyllm "github.com/mozilla-ai/any-llm-go/providers"
)

// RunProviderContract verifies a provider handles Tools + ToolChoice.
// The model must call the specified tool and return structured output.
func RunProviderContract(t *testing.T, p anyllm.Provider, model string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := p.Completion(ctx, anyllm.CompletionParams{
		Model: model,
		Messages: []anyllm.Message{
			{Role: "user", Content: "Fix the unused import: fmt is imported but not used in main.go"},
		},
		Tools: []anyllm.Tool{
			{
				Type: "function",
				Function: anyllm.Function{
					Name:        "apply_fix",
					Description: "Apply a code fix to a file",
					Parameters: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"file":    map[string]any{"type": "string", "description": "file path"},
							"content": map[string]any{"type": "string", "description": "complete file content"},
						},
						"required": []any{"file", "content"},
					},
				},
			},
		},
		ToolChoice: anyllm.ToolChoice{
			Type:     "function",
			Function: &anyllm.ToolChoiceFunction{Name: "apply_fix"},
		},
	})
	if err != nil {
		t.Fatalf("Completion with tools: %v", err)
	}

	if len(resp.Choices) == 0 {
		t.Fatal("no choices in response")
	}

	choice := resp.Choices[0]
	if len(choice.Message.ToolCalls) == 0 {
		t.Fatalf("expected tool_use response, got text: %v", choice.Message.Content)
	}

	tc := choice.Message.ToolCalls[0]
	if tc.Function.Name != "apply_fix" {
		t.Errorf("tool name = %q, want apply_fix", tc.Function.Name)
	}

	var args map[string]string
	if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
		t.Fatalf("parse tool args: %v (raw: %s)", err, tc.Function.Arguments)
	}

	if args["file"] == "" {
		t.Error("tool arg 'file' is empty")
	}
	if args["content"] == "" {
		t.Error("tool arg 'content' is empty")
	}

	t.Logf("Tool call: %s(file=%q, content_len=%d)", tc.Function.Name, args["file"], len(args["content"]))
}

func TestProviderContract_Vertex(t *testing.T) {
	region := os.Getenv("CLOUD_ML_REGION")
	project := os.Getenv("ANTHROPIC_VERTEX_PROJECT_ID")
	if region == "" || project == "" {
		t.Skip("Vertex credentials not configured")
	}

	p, err := NewVertexProvider(context.Background(), region, project)
	if err != nil {
		t.Fatalf("NewVertexProvider: %v", err)
	}

	RunProviderContract(t, p, "claude-sonnet-4-6")
}

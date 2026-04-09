package execution

import (
	"testing"

	anyllm "github.com/mozilla-ai/any-llm-go/providers"
)

func TestConvertMessages_NilContent(t *testing.T) {
	msgs := []anyllm.Message{
		{Role: "user", Content: nil},
	}
	out, _ := convertAllMessages(VertexConverter{}, msgs)
	if len(out) != 1 {
		t.Fatalf("messages = %d, want 1", len(out))
	}
	// No panic on nil content
}

func TestConvertMessages_NonStringContent(t *testing.T) {
	msgs := []anyllm.Message{
		{Role: "user", Content: 42}, // int instead of string
	}
	out, _ := convertAllMessages(VertexConverter{}, msgs)
	if len(out) != 1 {
		t.Fatalf("messages = %d, want 1", len(out))
	}
	// Type assertion to string fails gracefully (empty string), no panic
}

func TestConvertMessages_EmptyToolCallID(t *testing.T) {
	msgs := []anyllm.Message{
		{Role: "tool", ToolCallID: "", Content: "result"},
	}
	out, _ := convertAllMessages(VertexConverter{}, msgs)
	if len(out) != 1 {
		t.Fatalf("messages = %d, want 1", len(out))
	}
	// Empty tool_call_id still produces a message, no panic
}

func TestConvertMessages_MalformedToolCallArgs(t *testing.T) {
	msgs := []anyllm.Message{
		{
			Role: "assistant",
			ToolCalls: []anyllm.ToolCall{
				{ID: "c1", Type: "function", Function: anyllm.FunctionCall{
					Name:      "tool",
					Arguments: "not valid json{{{",
				}},
			},
		},
	}
	out, _ := convertAllMessages(VertexConverter{}, msgs)
	if len(out) != 1 {
		t.Fatalf("messages = %d, want 1", len(out))
	}
	// Malformed JSON in tool args → nil input, no panic
}

func TestConvertMessages_EmptyToolCalls(t *testing.T) {
	msgs := []anyllm.Message{
		{
			Role:      "assistant",
			Content:   "text only",
			ToolCalls: []anyllm.ToolCall{}, // empty slice, not nil
		},
	}
	out, _ := convertAllMessages(VertexConverter{}, msgs)
	if len(out) != 1 {
		t.Fatalf("messages = %d, want 1", len(out))
	}
	// Empty ToolCalls treated as text-only assistant message
}

func TestConvertMessages_ToolResultWithIsError(t *testing.T) {
	// Tool results don't carry isError through anyllm — verify no crash
	msgs := []anyllm.Message{
		{Role: "tool", ToolCallID: "c1", Content: "error: file not found"},
	}
	out, _ := convertAllMessages(VertexConverter{}, msgs)
	if len(out) != 1 {
		t.Fatalf("messages = %d, want 1", len(out))
	}
}

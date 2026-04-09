package execution

import (
	"testing"

	anyllm "github.com/mozilla-ai/any-llm-go/providers"
)

// RunConverterSecuritySuite verifies a MessageConverter handles malformed
// input without panicking. Every converter implementation must pass this.
func RunConverterSecuritySuite(t *testing.T, conv MessageConverter) {
	t.Helper()

	t.Run("NilContent", func(t *testing.T) {
		msgs := []anyllm.Message{{Role: "user", Content: nil}}
		out, _ := convertAllMessages(conv, msgs)
		if len(out) != 1 {
			t.Fatalf("messages = %d, want 1", len(out))
		}
	})

	t.Run("NonStringContent", func(t *testing.T) {
		msgs := []anyllm.Message{{Role: "user", Content: 42}}
		out, _ := convertAllMessages(conv, msgs)
		if len(out) != 1 {
			t.Fatalf("messages = %d, want 1", len(out))
		}
	})

	t.Run("EmptyToolCallID", func(t *testing.T) {
		msgs := []anyllm.Message{{Role: "tool", ToolCallID: "", Content: "result"}}
		out, _ := convertAllMessages(conv, msgs)
		if len(out) != 1 {
			t.Fatalf("messages = %d, want 1", len(out))
		}
	})

	t.Run("MalformedToolCallArgs", func(t *testing.T) {
		msgs := []anyllm.Message{{
			Role: "assistant",
			ToolCalls: []anyllm.ToolCall{{
				ID: "c1", Type: "function",
				Function: anyllm.FunctionCall{Name: "tool", Arguments: "not valid json{{{"},
			}},
		}}
		out, _ := convertAllMessages(conv, msgs)
		if len(out) != 1 {
			t.Fatalf("messages = %d, want 1", len(out))
		}
	})

	t.Run("EmptyToolCalls", func(t *testing.T) {
		msgs := []anyllm.Message{{
			Role:      "assistant",
			Content:   "text only",
			ToolCalls: []anyllm.ToolCall{},
		}}
		out, _ := convertAllMessages(conv, msgs)
		if len(out) != 1 {
			t.Fatalf("messages = %d, want 1", len(out))
		}
	})

	t.Run("ToolResultErrorContent", func(t *testing.T) {
		msgs := []anyllm.Message{{Role: "tool", ToolCallID: "c1", Content: "error: file not found"}}
		out, _ := convertAllMessages(conv, msgs)
		if len(out) != 1 {
			t.Fatalf("messages = %d, want 1", len(out))
		}
	})
}

// Every converter runs the security suite.

func TestConverterSecurity_Vertex(t *testing.T) {
	RunConverterSecuritySuite(t, VertexConverter{})
}

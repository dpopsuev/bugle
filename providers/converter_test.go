package providers

import (
	"testing"

	anyllm "github.com/mozilla-ai/any-llm-go/providers"
)

func TestConvertMessages_UserOnly(t *testing.T) {
	out, system := convertMessages([]anyllm.Message{{Role: "user", Content: "hello"}})
	if system != "" {
		t.Fatalf("system = %q, want empty", system)
	}
	if len(out) != 1 {
		t.Fatalf("messages = %d, want 1", len(out))
	}
}

func TestConvertMessages_SystemExtracted(t *testing.T) {
	out, system := convertMessages([]anyllm.Message{
		{Role: "system", Content: "respond in French"},
		{Role: "user", Content: "hello"},
	})
	if system != "respond in French" {
		t.Fatalf("system = %q", system)
	}
	if len(out) != 1 {
		t.Fatalf("messages = %d, want 1 (system extracted)", len(out))
	}
}

func TestConvertMessages_ToolResult(t *testing.T) {
	out, _ := convertMessages([]anyllm.Message{
		{Role: "tool", ToolCallID: "call-1", Content: "4"},
	})
	if len(out) != 1 {
		t.Fatalf("messages = %d, want 1", len(out))
	}
}

func TestConvertMessages_AssistantWithToolCalls(t *testing.T) {
	out, _ := convertMessages([]anyllm.Message{{
		Role:    "assistant",
		Content: "I'll calculate",
		ToolCalls: []anyllm.ToolCall{
			{ID: "c1", Type: "function", Function: anyllm.FunctionCall{Name: "calc", Arguments: `{"expr":"2+2"}`}},
		},
	}})
	if len(out) != 1 {
		t.Fatalf("messages = %d, want 1", len(out))
	}
}

func TestConvertMessages_FullRoundTrip(t *testing.T) {
	out, system := convertMessages([]anyllm.Message{
		{Role: "system", Content: "be helpful"},
		{Role: "user", Content: "what is 2+2?"},
		{
			Role:    "assistant",
			Content: "I'll calculate",
			ToolCalls: []anyllm.ToolCall{
				{ID: "c1", Type: "function", Function: anyllm.FunctionCall{Name: "calc", Arguments: `{"expr":"2+2"}`}},
			},
		},
		{Role: "tool", ToolCallID: "c1", Content: "4"},
	})
	if system != "be helpful" {
		t.Fatalf("system = %q", system)
	}
	if len(out) != 3 {
		t.Fatalf("messages = %d, want 3 (user + assistant + tool_result)", len(out))
	}
}

func TestConvertMessages_EmptyContent(t *testing.T) {
	out, _ := convertMessages([]anyllm.Message{{Role: "user", Content: ""}})
	if len(out) != 1 {
		t.Fatalf("messages = %d, want 1", len(out))
	}
}

func TestConvertMessages_UnknownRole(t *testing.T) {
	out, _ := convertMessages([]anyllm.Message{
		{Role: "custom", Content: "ignored"},
		{Role: "user", Content: "kept"},
	})
	if len(out) != 1 {
		t.Fatalf("messages = %d, want 1 (unknown role skipped with warning)", len(out))
	}
}

func TestConvertMessages_MultipleSystemMessages(t *testing.T) {
	_, system := convertMessages([]anyllm.Message{
		{Role: "system", Content: "rule 1"},
		{Role: "system", Content: "rule 2"},
		{Role: "user", Content: "hello"},
	})
	if system != "rule 1\nrule 2" {
		t.Fatalf("system = %q, want 'rule 1\\nrule 2'", system)
	}
}

// --- Security: malformed input must not panic ---

func TestConvertMessages_NilContent(t *testing.T) {
	out, _ := convertMessages([]anyllm.Message{{Role: "user", Content: nil}})
	if len(out) != 1 {
		t.Fatalf("messages = %d, want 1", len(out))
	}
}

func TestConvertMessages_NonStringContent(t *testing.T) {
	out, _ := convertMessages([]anyllm.Message{{Role: "user", Content: 42}})
	if len(out) != 1 {
		t.Fatalf("messages = %d, want 1", len(out))
	}
}

func TestConvertMessages_EmptyToolCallID(t *testing.T) {
	out, _ := convertMessages([]anyllm.Message{{Role: "tool", ToolCallID: "", Content: "result"}})
	if len(out) != 1 {
		t.Fatalf("messages = %d, want 1", len(out))
	}
}

func TestConvertMessages_MalformedToolCallArgs(t *testing.T) {
	out, _ := convertMessages([]anyllm.Message{{
		Role: "assistant",
		ToolCalls: []anyllm.ToolCall{{
			ID: "c1", Type: "function",
			Function: anyllm.FunctionCall{Name: "tool", Arguments: "not valid json{{{"},
		}},
	}})
	if len(out) != 1 {
		t.Fatalf("messages = %d, want 1", len(out))
	}
}

func TestConvertMessages_EmptyToolCallsSlice(t *testing.T) {
	out, _ := convertMessages([]anyllm.Message{{
		Role:      "assistant",
		Content:   "text only",
		ToolCalls: []anyllm.ToolCall{},
	}})
	if len(out) != 1 {
		t.Fatalf("messages = %d, want 1", len(out))
	}
}

func TestConvertMessages_ToolResultErrorContent(t *testing.T) {
	out, _ := convertMessages([]anyllm.Message{{Role: "tool", ToolCallID: "c1", Content: "error: file not found"}})
	if len(out) != 1 {
		t.Fatalf("messages = %d, want 1", len(out))
	}
}

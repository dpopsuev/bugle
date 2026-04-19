// converter.go — converts anyllm.Messages to Anthropic SDK types.
//
// Plain functions, no interface. The switch has a default that warns
// on unknown roles instead of silently dropping them (TRP-BUG-2 fix).
package providers

import (
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"

	anyllm "github.com/mozilla-ai/any-llm-go/providers"
)

// convertMessages converts anyllm messages to Anthropic SDK format.
// Returns the converted messages and the combined system prompt.
// Unknown roles are logged and skipped — never silently dropped.
func convertMessages(msgs []anyllm.Message) ([]anthropic.MessageParam, string) {
	out := make([]anthropic.MessageParam, 0, len(msgs))
	var systemParts []string

	for _, m := range msgs {
		switch m.Role {
		case "system":
			content, _ := m.Content.(string)
			systemParts = append(systemParts, content)

		case vertexRoleUser:
			content, _ := m.Content.(string)
			out = append(out, anthropic.NewUserMessage(anthropic.NewTextBlock(content)))

		case vertexRoleAssistant:
			out = append(out, convertAssistantMessage(m))

		case "tool":
			content, _ := m.Content.(string)
			out = append(out, anthropic.NewUserMessage(
				anthropic.NewToolResultBlock(m.ToolCallID, content, false),
			))

		default:
			slog.Warn("unknown message role, skipping",
				slog.String("role", m.Role),
			)
		}
	}

	return out, strings.Join(systemParts, "\n")
}

// convertAssistantMessage handles assistant messages, preserving tool_use blocks.
func convertAssistantMessage(msg anyllm.Message) anthropic.MessageParam {
	if len(msg.ToolCalls) == 0 {
		content, _ := msg.Content.(string)
		return anthropic.NewAssistantMessage(anthropic.NewTextBlock(content))
	}

	blocks := make([]anthropic.ContentBlockParamUnion, 0, len(msg.ToolCalls)+1)
	if content, _ := msg.Content.(string); content != "" {
		blocks = append(blocks, anthropic.NewTextBlock(content))
	}
	for _, tc := range msg.ToolCalls {
		var input map[string]any
		_ = json.Unmarshal([]byte(tc.Function.Arguments), &input)
		blocks = append(blocks, anthropic.ContentBlockParamUnion{
			OfToolUse: &anthropic.ToolUseBlockParam{
				Type:  "tool_use",
				ID:    tc.ID,
				Name:  tc.Function.Name,
				Input: input,
			},
		})
	}
	return anthropic.NewAssistantMessage(blocks...)
}

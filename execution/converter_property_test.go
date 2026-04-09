package execution

import (
	"testing"
	"testing/quick"

	anyllm "github.com/mozilla-ai/any-llm-go/providers"
)

// Property: output count = input - system - unknown roles.
func TestConvertMessages_Property_PreservesCount(t *testing.T) {
	f := func(nUser, nAssistant, nSystem, nTool, nUnknown uint8) bool {
		nu, na, ns, nt, nk := int(nUser%5), int(nAssistant%5), int(nSystem%5), int(nTool%5), int(nUnknown%5)

		var msgs []anyllm.Message
		for range nu {
			msgs = append(msgs, anyllm.Message{Role: "user", Content: "u"})
		}
		for range na {
			msgs = append(msgs, anyllm.Message{Role: "assistant", Content: "a"})
		}
		for range ns {
			msgs = append(msgs, anyllm.Message{Role: "system", Content: "s"})
		}
		for range nt {
			msgs = append(msgs, anyllm.Message{Role: "tool", ToolCallID: "c1", Content: "r"})
		}
		for range nk {
			msgs = append(msgs, anyllm.Message{Role: "custom", Content: "x"})
		}

		out, _ := convertMessages(msgs)
		expected := nu + na + nt
		return len(out) == expected
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 200}); err != nil {
		t.Fatal(err)
	}
}

// Property: system content always extracted, never in message list.
func TestConvertMessages_Property_SystemAlwaysExtracted(t *testing.T) {
	f := func(systemText string) bool {
		if systemText == "" {
			return true
		}
		msgs := []anyllm.Message{
			{Role: "system", Content: systemText},
			{Role: "user", Content: "hello"},
		}
		out, system := convertMessages(msgs)
		return system != "" && len(out) == 1
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Fatal(err)
	}
}

// Property: tool results never dropped.
func TestConvertMessages_Property_ToolResultNeverDropped(t *testing.T) {
	f := func(nTool uint8) bool {
		n := int(nTool%10) + 1
		var msgs []anyllm.Message
		for i := range n {
			msgs = append(msgs, anyllm.Message{
				Role:       "tool",
				ToolCallID: string(rune('a' + i)),
				Content:    "result",
			})
		}
		out, _ := convertMessages(msgs)
		return len(out) == n
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Fatal(err)
	}
}

package transport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/a2aproject/a2a-go/a2a"
	"github.com/a2aproject/a2a-go/a2aclient"
	"github.com/a2aproject/a2a-go/a2aclient/agentcard"
)

func testCard(url string) a2a.AgentCard {
	return a2a.AgentCard{
		Name:               "test-agent",
		Version:            "1.0.0",
		ProtocolVersion:    "1.0",
		PreferredTransport: a2a.TransportProtocolJSONRPC,
		URL:                url,
		DefaultInputModes:  []string{"text/plain"},
		DefaultOutputModes: []string{"text/plain"},
		Skills: []a2a.AgentSkill{{
			ID:   "echo",
			Name: "echo",
		}},
	}
}

func TestA2AServer_MessageSendRoundTrip(t *testing.T) {
	tr := NewA2ATransport(testCard("http://localhost"))
	defer tr.Close()

	_ = tr.Register("agent-1", func(_ context.Context, msg Message) (Message, error) {
		return Message{Content: "echo: " + msg.Content, Performative: "inform"}, nil
	})

	ts := httptest.NewServer(tr.Mux())
	defer ts.Close()

	card := testCard(ts.URL)
	client, err := a2aclient.NewFromCard(context.Background(), &card, a2aclient.WithJSONRPCTransport(http.DefaultClient))
	if err != nil {
		t.Fatalf("NewFromCard: %v", err)
	}

	result, err := client.SendMessage(context.Background(), &a2a.MessageSendParams{
		Message: a2a.NewMessage(a2a.MessageRoleUser, &a2a.TextPart{Text: "hello"}),
	})
	if err != nil {
		t.Fatalf("SendMessage: %v", err)
	}

	task, ok := result.(*a2a.Task)
	if !ok {
		t.Fatalf("result type = %T, want *a2a.Task", result)
	}

	if task.Status.State != a2a.TaskStateCompleted {
		t.Fatalf("task state = %s, want completed", task.Status.State)
	}

	t.Logf("Task status: state=%s", task.Status.State)
	if task.Status.Message != nil {
		t.Logf("Status message role=%s parts=%d", task.Status.Message.Role, len(task.Status.Message.Parts))
		for i, part := range task.Status.Message.Parts {
			t.Logf("  part[%d] type=%T", i, part)
			if tp, ok := part.(*a2a.TextPart); ok {
				t.Logf("  part[%d] text=%q", i, tp.Text)
			}
		}
	} else {
		t.Log("Status message is nil")
	}

	if len(task.History) > 0 {
		t.Logf("History: %d messages", len(task.History))
		for i, m := range task.History {
			t.Logf("  history[%d] role=%s parts=%d", i, m.Role, len(m.Parts))
		}
	}

	var content string
	if task.Status.Message != nil {
		content = extractText(task.Status.Message.Parts)
	}
	if content == "" && len(task.History) > 0 {
		content = extractText(task.History[len(task.History)-1].Parts)
	}

	if content == "" {
		t.Fatal("no content in status message or history")
	}
	t.Logf("A2A round-trip: %s", content)
}

func TestA2AServer_AgentCardDiscovery(t *testing.T) {
	tr := NewA2ATransport(testCard("http://localhost"))
	defer tr.Close()

	ts := httptest.NewServer(tr.Mux())
	defer ts.Close()

	resolver := agentcard.NewResolver(http.DefaultClient)
	resolved, err := resolver.Resolve(context.Background(), ts.URL)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	if resolved.Name != "test-agent" {
		t.Fatalf("name = %q, want test-agent", resolved.Name)
	}
	if resolved.ProtocolVersion != "1.0" {
		t.Fatalf("protocol = %q, want 1.0", resolved.ProtocolVersion)
	}
	if len(resolved.Skills) != 1 {
		t.Fatalf("skills = %d, want 1", len(resolved.Skills))
	}
}

func extractText(parts a2a.ContentParts) string {
	for _, part := range parts {
		switch tp := part.(type) {
		case *a2a.TextPart:
			return tp.Text
		case a2a.TextPart:
			return tp.Text
		}
	}
	return ""
}

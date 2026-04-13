package transport

import (
	"context"
	"errors"
	"testing"
)

var errNoEdge = errors.New("no communicates_with edge")

func TestRouteGuard_BlocksWithoutEdge(t *testing.T) {
	tr := NewLocalTransport()
	defer tr.Close()

	_ = tr.Register("agent-a", func(_ context.Context, msg Message) (Message, error) {
		return Message{From: "agent-a", Content: "reply"}, nil
	})

	// Set guard that rejects all messages.
	tr.SetRouteGuard(func(_, _ AgentID) error {
		return errNoEdge
	})

	_, err := tr.SendMessage(context.Background(), "agent-a", Message{From: "agent-b", Content: "hello"})
	if err == nil {
		t.Fatal("expected route guard rejection")
	}
	if !errors.Is(err, errNoEdge) {
		t.Fatalf("error = %v, want errNoEdge", err)
	}
}

func TestRouteGuard_AllowsWithEdge(t *testing.T) {
	tr := NewLocalTransport()
	defer tr.Close()

	_ = tr.Register("agent-a", func(_ context.Context, msg Message) (Message, error) {
		return Message{From: "agent-a", Content: "allowed"}, nil
	})

	// Guard that allows all.
	tr.SetRouteGuard(func(_, _ AgentID) error { return nil })

	resp, err := tr.Ask(context.Background(), "agent-a", Message{From: "agent-b", Content: "hello"})
	if err != nil {
		t.Fatalf("expected success: %v", err)
	}
	if resp.Content != "allowed" {
		t.Errorf("content = %q", resp.Content)
	}
}

func TestRouteGuard_NilGuardAllowsAll(t *testing.T) {
	tr := NewLocalTransport()
	defer tr.Close()

	_ = tr.Register("agent-a", func(_ context.Context, msg Message) (Message, error) {
		return Message{From: "agent-a", Content: "no guard"}, nil
	})

	// No guard set — everything allowed.
	resp, err := tr.Ask(context.Background(), "agent-a", Message{From: "agent-b"})
	if err != nil {
		t.Fatalf("expected success: %v", err)
	}
	if resp.Content != "no guard" {
		t.Errorf("content = %q", resp.Content)
	}
}

func TestRouteGuard_SelectiveEdgeCheck(t *testing.T) {
	tr := NewLocalTransport()
	defer tr.Close()

	_ = tr.Register("allowed-target", func(_ context.Context, _ Message) (Message, error) {
		return Message{Content: "ok"}, nil
	})
	_ = tr.Register("blocked-target", func(_ context.Context, _ Message) (Message, error) {
		return Message{Content: "should not reach"}, nil
	})

	// Guard that only allows messages to "allowed-target".
	tr.SetRouteGuard(func(_, to AgentID) error {
		if to == "allowed-target" {
			return nil
		}
		return errNoEdge
	})

	// Allowed target works.
	_, err := tr.Ask(context.Background(), "allowed-target", Message{From: "sender"})
	if err != nil {
		t.Fatalf("allowed target should work: %v", err)
	}

	// Blocked target rejected.
	_, err = tr.SendMessage(context.Background(), "blocked-target", Message{From: "sender"})
	if !errors.Is(err, errNoEdge) {
		t.Fatalf("blocked target should be rejected: %v", err)
	}
}

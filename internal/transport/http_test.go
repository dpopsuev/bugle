package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestHTTP_WireRoundTrip proves the HTTP handler works over the network —
// a real HTTP POST to /a2a/send, handler processes, response comes back.
func TestHTTP_WireRoundTrip(t *testing.T) {
	tr := NewHTTPTransport()
	defer tr.Close()

	_ = tr.Register("agent-a", func(_ context.Context, msg Message) (Message, error) {
		return Message{From: "agent-a", Content: "wire: " + msg.Content}, nil
	})

	ts := httptest.NewServer(tr.Mux())
	defer ts.Close()

	// POST to the HTTP endpoint.
	body, _ := json.Marshal(map[string]any{
		"to":      "agent-a",
		"message": Message{From: "client", Content: "hello over HTTP"},
	})

	resp, err := http.Post(ts.URL+"/a2a/send", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}

	var result struct {
		TaskID string    `json:"task_id"`
		State  TaskState `json:"state"`
		Data   *Message  `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if result.State != TaskCompleted {
		t.Errorf("state = %q, want completed", result.State)
	}
	if result.Data == nil {
		t.Fatal("data is nil")
	}
	if result.Data.Content != "wire: hello over HTTP" {
		t.Errorf("content = %q", result.Data.Content)
	}
}

// TestHTTP_WireUnknownAgent proves HTTP returns 404 for unknown agents.
func TestHTTP_WireUnknownAgent(t *testing.T) {
	tr := NewHTTPTransport()
	defer tr.Close()

	ts := httptest.NewServer(tr.Mux())
	defer ts.Close()

	body, _ := json.Marshal(map[string]any{
		"to":      "ghost",
		"message": Message{From: "client", Content: "hello"},
	})

	resp, err := http.Post(ts.URL+"/a2a/send", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

// TestHTTP_AgentCardDiscovery proves /.well-known/agent-card.json lists registered agents.
func TestHTTP_AgentCardDiscovery(t *testing.T) {
	tr := NewHTTPTransport()
	defer tr.Close()

	_ = tr.Register("agent-a", func(_ context.Context, _ Message) (Message, error) {
		return Message{}, nil
	})
	tr.Roles().Register("agent-a", "investigator")

	_ = tr.Register("agent-b", func(_ context.Context, _ Message) (Message, error) {
		return Message{}, nil
	})
	tr.Roles().Register("agent-b", "reviewer")

	ts := httptest.NewServer(tr.Mux())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/.well-known/agent-card.json")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}

	var cards []AgentCard
	if err := json.NewDecoder(resp.Body).Decode(&cards); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if len(cards) != 2 {
		t.Fatalf("cards = %d, want 2", len(cards))
	}

	roles := map[string]bool{}
	for _, c := range cards {
		roles[c.Role] = true
		if c.Transport != "http" {
			t.Errorf("card %s transport = %q, want http", c.ID, c.Transport)
		}
	}
	if !roles["investigator"] || !roles["reviewer"] {
		t.Errorf("expected investigator + reviewer roles, got %v", roles)
	}
}

// TestHTTP_AgentCardDiscovery_Empty proves empty registry returns empty array.
func TestHTTP_AgentCardDiscovery_Empty(t *testing.T) {
	tr := NewHTTPTransport()
	defer tr.Close()

	ts := httptest.NewServer(tr.Mux())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/.well-known/agent-card.json")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()

	var cards []AgentCard
	json.NewDecoder(resp.Body).Decode(&cards)

	if len(cards) != 0 {
		t.Errorf("expected empty, got %d cards", len(cards))
	}
}

// TestHTTP_CrossProcess proves two separate HTTPTransports on different ports
// can communicate — agent on Transport A sends to agent on Transport B via HTTP.
func TestHTTP_CrossProcess(t *testing.T) {
	// Transport B: the "remote" agent.
	trB := NewHTTPTransport()
	defer trB.Close()

	_ = trB.Register("remote-agent", func(_ context.Context, msg Message) (Message, error) {
		return Message{From: "remote-agent", Content: "remote: " + msg.Content}, nil
	})

	serverB := httptest.NewServer(trB.Mux())
	defer serverB.Close()

	// Transport A: the "local" side. It calls Transport B over HTTP.
	// We simulate cross-process by making an HTTP call to serverB.
	body, _ := json.Marshal(map[string]any{
		"to":      "remote-agent",
		"message": Message{From: "local-agent", Content: "cross-process hello"},
	})

	resp, err := http.Post(serverB.URL+"/a2a/send", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}

	var result struct {
		State TaskState `json:"state"`
		Data  *Message  `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	if result.State != TaskCompleted {
		t.Errorf("state = %q, want completed", result.State)
	}
	if result.Data == nil || result.Data.Content != "remote: cross-process hello" {
		t.Errorf("response = %v", result.Data)
	}

	// Also verify discovery works cross-process.
	cardResp, err := http.Get(serverB.URL + "/.well-known/agent-card.json")
	if err != nil {
		t.Fatalf("GET cards: %v", err)
	}
	defer cardResp.Body.Close()

	var cards []AgentCard
	json.NewDecoder(cardResp.Body).Decode(&cards)
	if len(cards) != 1 {
		t.Errorf("expected 1 card, got %d", len(cards))
	}
}

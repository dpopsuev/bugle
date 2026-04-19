package arsenal

import (
	"context"
	"os"
	"testing"
)

func TestExtractProvider(t *testing.T) {
	tests := []struct {
		id   string
		want string
	}{
		{"anthropic/claude-sonnet-4", "anthropic"},
		{"openai/gpt-4o", "openai"},
		{"google/gemini-2.5-pro", "google"},
		{"mistralai/mistral-large", "mistralai"},
		{"bare-model", "openrouter"},
	}
	for _, tt := range tests {
		got := extractProvider(tt.id)
		if got != tt.want {
			t.Errorf("extractProvider(%q) = %q, want %q", tt.id, got, tt.want)
		}
	}
}

func TestOpenRouterDiscoverer_Interface(t *testing.T) {
	var _ ModelDiscoverer = (*OpenRouterDiscoverer)(nil)
}

func TestOpenRouterDiscoverer_E2E(t *testing.T) {
	if os.Getenv("TROUPE_TEST_LIVE_LLM") == "" {
		t.Skip("TROUPE_TEST_LIVE_LLM not set — skipping network test")
	}

	d := NewOpenRouterDiscoverer()
	models, err := d.Discover(context.Background())
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if len(models) < 50 {
		t.Errorf("expected 50+ models, got %d", len(models))
	}

	var hasAnthropic, hasOpenAI bool
	for _, m := range models {
		switch m.Provider {
		case "anthropic":
			hasAnthropic = true
		case "openai":
			hasOpenAI = true
		}
	}
	if !hasAnthropic {
		t.Error("expected anthropic models")
	}
	if !hasOpenAI {
		t.Error("expected openai models")
	}

	t.Logf("OpenRouter discovered %d models", len(models))
}

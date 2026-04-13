package arsenal

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestAnthropicDiscoverer_NilWithoutKey(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")
	d := NewAnthropicDiscoverer()
	if d != nil {
		t.Fatal("should return nil without API key")
	}
}

func TestAnthropicDiscoverer_RealAPI(t *testing.T) {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	d := NewAnthropicDiscoverer()
	if d == nil {
		t.Fatal("discoverer should not be nil with key set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) //nolint:mnd // API timeout
	defer cancel()

	models, err := d.Discover(ctx)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}

	if len(models) == 0 {
		t.Fatal("expected at least one model from Anthropic API")
	}

	// All returned models should be available and from anthropic.
	for _, m := range models {
		if m.Provider != "anthropic" {
			t.Errorf("model %s: provider = %q", m.ID, m.Provider)
		}
		if !m.Available {
			t.Errorf("model %s: should be available (listed by API)", m.ID)
		}
	}

	t.Logf("discovered %d Anthropic models", len(models))
	for _, m := range models {
		t.Logf("  %s", m.ID)
	}
}

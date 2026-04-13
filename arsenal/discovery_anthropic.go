// discovery_anthropic.go — Anthropic model discovery via GET /v1/models.
//
// Uses anthropic-sdk-go to fetch the live model catalog.
// Models found in the API are marked Available=true.
// For Vertex users, a rawPredict probe could verify per-project access
// (deferred — Vertex has no models list for partner models).
//
// TRP-TSK-36, TRP-GOL-8
package arsenal

import (
	"context"
	"os"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

var _ ModelDiscoverer = (*AnthropicDiscoverer)(nil)

// AnthropicDiscoverer fetches the Anthropic model catalog via /v1/models.
type AnthropicDiscoverer struct {
	client anthropic.Client
}

// NewAnthropicDiscoverer creates a discoverer using the ANTHROPIC_API_KEY env var.
// Returns nil if no API key is configured.
func NewAnthropicDiscoverer() *AnthropicDiscoverer {
	key := os.Getenv("ANTHROPIC_API_KEY")
	if key == "" {
		return nil
	}
	return &AnthropicDiscoverer{
		client: anthropic.NewClient(option.WithAPIKey(key)),
	}
}

func (d *AnthropicDiscoverer) Provider() string { return "anthropic" }

func (d *AnthropicDiscoverer) Discover(ctx context.Context) ([]DiscoveredModel, error) {
	pager := d.client.Models.ListAutoPaging(ctx, anthropic.ModelListParams{})

	var models []DiscoveredModel
	for pager.Next() {
		info := pager.Current()
		models = append(models, DiscoveredModel{
			ID:        info.ID,
			Provider:  "anthropic",
			Available: true, // listed by API = available
		})
	}
	if err := pager.Err(); err != nil {
		return models, err
	}

	return models, nil
}

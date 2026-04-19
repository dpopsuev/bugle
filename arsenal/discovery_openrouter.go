package arsenal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

var _ ModelDiscoverer = (*OpenRouterDiscoverer)(nil)

// OpenRouterDiscoverer fetches the OpenRouter model catalog via
// GET /api/v1/models. Free, no auth required.
type OpenRouterDiscoverer struct {
	client *http.Client
}

// NewOpenRouterDiscoverer creates a discoverer for OpenRouter's free
// model listing API.
func NewOpenRouterDiscoverer() *OpenRouterDiscoverer {
	return &OpenRouterDiscoverer{client: &http.Client{}}
}

func (d *OpenRouterDiscoverer) Provider() string { return "openrouter" }

func (d *OpenRouterDiscoverer) Discover(ctx context.Context) ([]DiscoveredModel, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://openrouter.ai/api/v1/models", http.NoBody)
	if err != nil {
		return nil, err
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openrouter models list: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: openrouter status %d", ErrDiscoveryAPI, resp.StatusCode)
	}

	var body struct {
		Data []openRouterModel `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("openrouter models decode: %w", err)
	}

	models := make([]DiscoveredModel, 0, len(body.Data))
	for _, m := range body.Data {
		models = append(models, DiscoveredModel{
			ID:            m.ID,
			Provider:      extractProvider(m.ID),
			ContextWindow: m.ContextLength,
			Available:     true,
			Capabilities:  m.SupportedParameters,
		})
	}

	return models, nil
}

type openRouterModel struct {
	ID                  string   `json:"id"`
	Name                string   `json:"name"`
	ContextLength       int      `json:"context_length"`
	SupportedParameters []string `json:"supported_parameters"`
}

func extractProvider(modelID string) string {
	if idx := strings.IndexByte(modelID, '/'); idx > 0 {
		return modelID[:idx]
	}
	return "openrouter"
}

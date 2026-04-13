// discovery_gemini.go — Gemini model discovery via GET /v1beta/models.
//
// Raw HTTP — no heavy Google Cloud SDK for a simple model list.
// Gemini returns rich metadata: context window, capabilities, versioning.
//
// TRP-TSK-38, TRP-GOL-8
package arsenal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

var _ ModelDiscoverer = (*GeminiDiscoverer)(nil)

// GeminiDiscoverer fetches the Gemini model catalog via /v1beta/models.
type GeminiDiscoverer struct {
	apiKey string
	client *http.Client
}

// NewGeminiDiscoverer creates a discoverer using GEMINI_API_KEY env var.
// Returns nil if no key is set.
func NewGeminiDiscoverer() *GeminiDiscoverer {
	key := os.Getenv("GEMINI_API_KEY")
	if key == "" {
		return nil
	}
	return &GeminiDiscoverer{apiKey: key, client: &http.Client{}}
}

func (d *GeminiDiscoverer) Provider() string { return "google" }

func (d *GeminiDiscoverer) Discover(ctx context.Context) ([]DiscoveredModel, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models?key=%s&pageSize=100", d.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gemini models list: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini models list: status %d", resp.StatusCode)
	}

	var body struct {
		Models []struct {
			Name                       string   `json:"name"`
			DisplayName                string   `json:"displayName"`
			InputTokenLimit            int      `json:"inputTokenLimit"`
			OutputTokenLimit           int      `json:"outputTokenLimit"`
			SupportedGenerationMethods []string `json:"supportedGenerationMethods"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("gemini models decode: %w", err)
	}

	var models []DiscoveredModel
	for _, m := range body.Models {
		// Filter to generative models (skip embedding models).
		if !supportsGeneration(m.SupportedGenerationMethods) {
			continue
		}

		// Strip "models/" prefix from name.
		id := strings.TrimPrefix(m.Name, "models/")

		models = append(models, DiscoveredModel{
			ID:            id,
			Provider:      "google",
			ContextWindow: m.InputTokenLimit,
			Available:     true,
			Capabilities:  m.SupportedGenerationMethods,
		})
	}

	return models, nil
}

func supportsGeneration(methods []string) bool {
	for _, m := range methods {
		if m == "generateContent" {
			return true
		}
	}
	return false
}

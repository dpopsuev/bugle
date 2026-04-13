// discovery_openai.go — OpenAI model discovery via GET /v1/models.
//
// Raw HTTP — no SDK import needed for a simple model list.
// OpenAI returns {id, owned_by} per model. No capabilities or pricing.
// Filter to chat-capable models by owned_by prefix.
//
// TRP-TSK-37, TRP-GOL-8
package arsenal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

var _ ModelDiscoverer = (*OpenAIDiscoverer)(nil)

// OpenAIDiscoverer fetches the OpenAI model catalog via /v1/models.
type OpenAIDiscoverer struct {
	apiKey string
	client *http.Client
}

// NewOpenAIDiscoverer creates a discoverer using OPENAI_API_KEY env var.
// Returns nil if no key is set.
func NewOpenAIDiscoverer() *OpenAIDiscoverer {
	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		return nil
	}
	return &OpenAIDiscoverer{apiKey: key, client: &http.Client{}}
}

func (d *OpenAIDiscoverer) Provider() string { return "openai" }

func (d *OpenAIDiscoverer) Discover(ctx context.Context) ([]DiscoveredModel, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.openai.com/v1/models", http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+d.apiKey)

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai models list: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai models list: status %d", resp.StatusCode)
	}

	var body struct {
		Data []struct {
			ID      string `json:"id"`
			OwnedBy string `json:"owned_by"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("openai models decode: %w", err)
	}

	var models []DiscoveredModel
	for _, m := range body.Data {
		// Filter to chat models — skip embeddings, whisper, dall-e, etc.
		if !isChatModel(m.ID) {
			continue
		}
		models = append(models, DiscoveredModel{
			ID:        m.ID,
			Provider:  "openai",
			Available: true,
		})
	}

	return models, nil
}

// isChatModel filters to models usable for chat completion.
func isChatModel(id string) bool {
	for _, prefix := range []string{"gpt-", "o1", "o3", "o4", "chatgpt"} {
		if strings.HasPrefix(id, prefix) {
			return true
		}
	}
	return false
}

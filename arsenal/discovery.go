// discovery.go — ModelDiscoverer interface for live model catalog discovery.
//
// Hexagonal port: Arsenal defines the contract, per-provider adapters implement it.
// StubDiscoverer ships for testkit (Forge rule: stub with implementation).
//
// TRP-TSK-128, TRP-GOL-8
package arsenal

import "context"

// ModelDiscoverer fetches a live model catalog from a provider API.
// Implementations must be safe for concurrent use.
type ModelDiscoverer interface {
	// Discover fetches the current model catalog from the provider.
	// Returns available models with metadata. Errors are non-fatal —
	// Arsenal falls back to static YAML when discovery fails.
	Discover(ctx context.Context) ([]DiscoveredModel, error)

	// Provider returns the provider name (e.g., "anthropic", "openai", "google").
	Provider() string
}

// DiscoveredModel is a model found via live API discovery.
type DiscoveredModel struct {
	ID            string   // model identifier (e.g., "claude-opus-4-6")
	Provider      string   // provider name
	ContextWindow int      // context window in tokens (0 if unknown)
	Available     bool     // true if the model responded to a probe or was listed by the API
	Capabilities  []string // provider-reported capabilities (e.g., "tool_use", "vision")
}

// StubDiscoverer is a test double that returns canned results.
type StubDiscoverer struct {
	ProviderName string
	Models       []DiscoveredModel
	Err          error
	// DiscoverCalls records how many times Discover was called.
	DiscoverCalls int
}

var _ ModelDiscoverer = (*StubDiscoverer)(nil)

func (s *StubDiscoverer) Discover(_ context.Context) ([]DiscoveredModel, error) {
	s.DiscoverCalls++
	return s.Models, s.Err
}

func (s *StubDiscoverer) Provider() string { return s.ProviderName }

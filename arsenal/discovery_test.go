package arsenal

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestStubDiscoverer_Contract(t *testing.T) {
	stub := &StubDiscoverer{
		ProviderName: "test-provider",
		Models: []DiscoveredModel{
			{ID: "model-a", Provider: "test-provider", Available: true, ContextWindow: 100000},
			{ID: "model-b", Provider: "test-provider", Available: false},
		},
	}

	// Provider returns name.
	if stub.Provider() != "test-provider" {
		t.Errorf("Provider = %q", stub.Provider())
	}

	// Discover returns models.
	models, err := stub.Discover(context.Background())
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if len(models) != 2 {
		t.Fatalf("models = %d, want 2", len(models))
	}
	if !models[0].Available {
		t.Error("model-a should be available")
	}
	if models[1].Available {
		t.Error("model-b should be unavailable")
	}
	if stub.DiscoverCalls != 1 {
		t.Errorf("DiscoverCalls = %d, want 1", stub.DiscoverCalls)
	}
}

func TestStubDiscoverer_Error(t *testing.T) {
	stub := &StubDiscoverer{
		ProviderName: "failing",
		Err:          errors.New("api unreachable"),
	}

	_, err := stub.Discover(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "api unreachable" {
		t.Errorf("error = %q", err.Error())
	}
}

func TestModelDiscoverer_InterfaceCompliance(t *testing.T) {
	var _ ModelDiscoverer = (*StubDiscoverer)(nil)
}

// TestDiscovery_E2E_MergeAndSelect proves the full pipeline:
// static catalog → live discovery → merge → Select() only returns available models.
func TestDiscovery_E2E_MergeAndSelect(t *testing.T) {
	a, err := NewArsenal("")
	if err != nil {
		t.Fatalf("NewArsenal: %v", err)
	}

	snap := a.snapshots[a.active]

	// Collect all anthropic models from the static catalog.
	var anthropicModels []string
	for id, m := range snap.Models {
		if m.Provider == "anthropic" {
			anthropicModels = append(anthropicModels, id)
		}
	}
	if len(anthropicModels) == 0 {
		t.Fatal("no anthropic models in catalog")
	}

	// Stub discoverer: mark first model available, rest unavailable.
	available := anthropicModels[0]
	discovered := make([]DiscoveredModel, 0, len(anthropicModels))
	for _, id := range anthropicModels {
		discovered = append(discovered, DiscoveredModel{
			ID:        id,
			Provider:  "anthropic",
			Available: id == available,
		})
	}

	// Merge discovery results into snapshot.
	MergeDiscovery(snap, discovered)

	// Verify: only the available model passes Select().
	result, err := a.Select("", &Preferences{
		Providers: Filter{Allow: []string{"anthropic"}},
	})
	if err != nil {
		t.Fatalf("Select: %v", err)
	}
	if result.Model != available {
		t.Errorf("Select returned %q, want %q (only available model)", result.Model, available)
	}
}

// TestMergeDiscovery_UnknownModel proves new models from API are added.
func TestMergeDiscovery_UnknownModel(t *testing.T) {
	snap := &Snapshot{
		Models: map[string]*ModelEntry{
			"existing": {ID: "existing", Provider: "test", Available: false},
		},
	}

	MergeDiscovery(snap, []DiscoveredModel{
		{ID: "existing", Provider: "test", Available: true},
		{ID: "brand-new", Provider: "test", Available: true, ContextWindow: 200000},
	})

	if !snap.Models["existing"].Available {
		t.Error("existing model should be marked available")
	}
	newModel, ok := snap.Models["brand-new"]
	if !ok {
		t.Fatal("brand-new model should be added")
	}
	if !newModel.Available {
		t.Error("brand-new should be available")
	}
	if newModel.Context != 200000 {
		t.Errorf("context = %d, want 200000", newModel.Context)
	}
}

// TestMergeDiscovery_Empty proves empty discovery is a no-op.
// TestArsenal_Discover_FanOut proves the startup pipeline fans out to multiple providers.
func TestArsenal_Discover_FanOut(t *testing.T) {
	a, err := NewArsenal("")
	if err != nil {
		t.Fatal(err)
	}

	stubA := &StubDiscoverer{
		ProviderName: "provider-a",
		Models: []DiscoveredModel{
			{ID: "model-a1", Provider: "provider-a", Available: true},
		},
	}
	stubB := &StubDiscoverer{
		ProviderName: "provider-b",
		Models: []DiscoveredModel{
			{ID: "model-b1", Provider: "provider-b", Available: true},
		},
	}

	a.RegisterDiscoverer(stubA)
	a.RegisterDiscoverer(stubB)

	errs := a.Discover(context.Background())
	if len(errs) != 0 {
		t.Fatalf("Discover errors: %v", errs)
	}

	if stubA.DiscoverCalls != 1 {
		t.Errorf("stubA called %d times", stubA.DiscoverCalls)
	}
	if stubB.DiscoverCalls != 1 {
		t.Errorf("stubB called %d times", stubB.DiscoverCalls)
	}
}

// TestArsenal_Discover_PartialFailure proves one failing discoverer doesn't block others.
func TestArsenal_Discover_PartialFailure(t *testing.T) {
	a, err := NewArsenal("")
	if err != nil {
		t.Fatal(err)
	}

	good := &StubDiscoverer{
		ProviderName: "good",
		Models:       []DiscoveredModel{{ID: "m1", Provider: "good", Available: true}},
	}
	bad := &StubDiscoverer{
		ProviderName: "bad",
		Err:          errors.New("api down"),
	}

	a.RegisterDiscoverer(good)
	a.RegisterDiscoverer(bad)

	errs := a.Discover(context.Background())
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	if !strings.Contains(errs[0].Error(), "bad") {
		t.Errorf("error should mention provider: %v", errs[0])
	}

	// Good discoverer's models should still be merged.
	snap := a.snapshots[a.active]
	if _, ok := snap.Models["m1"]; !ok {
		t.Error("good provider's model should be merged despite bad provider failing")
	}
}

// TestArsenal_Discover_NoDiscoverers is a no-op.
func TestArsenal_Discover_NoDiscoverers(t *testing.T) {
	a, err := NewArsenal("")
	if err != nil {
		t.Fatal(err)
	}

	errs := a.Discover(context.Background())
	if errs != nil {
		t.Errorf("expected nil errors, got %v", errs)
	}
}

// TestArsenal_RegisterDiscoverer_NilSafe proves nil discoverers are ignored.
func TestArsenal_RegisterDiscoverer_NilSafe(t *testing.T) {
	a, err := NewArsenal("")
	if err != nil {
		t.Fatal(err)
	}

	a.RegisterDiscoverer(nil)
	if len(a.discoverers) != 0 {
		t.Error("nil discoverer should not be registered")
	}
}

func TestMergeDiscovery_Empty(t *testing.T) {
	snap := &Snapshot{
		Models: map[string]*ModelEntry{
			"m1": {ID: "m1", Provider: "test"},
		},
	}

	MergeDiscovery(snap, nil)

	if snap.discoveryRan {
		t.Error("discoveryRan should be false for empty discovery")
	}
}

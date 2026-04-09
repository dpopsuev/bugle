package broker_test

import (
	"context"
	"testing"

	"github.com/dpopsuev/troupe"
	"github.com/dpopsuev/troupe/broker"
	"github.com/dpopsuev/troupe/world"
)

func TestNew_EmptyEndpoint_ReturnsLocal(t *testing.T) {
	b := broker.New("")
	if b == nil {
		t.Fatal("New(\"\") returned nil")
	}
}

func TestNew_HTTPSEndpoint_ReturnsRemote(t *testing.T) {
	b := broker.New("https://cluster:8080")
	if b == nil {
		t.Fatal("New(\"https://...\") returned nil")
	}
}

func TestDefaultBroker_Pick_DefaultCount(t *testing.T) {
	b := broker.New("")
	configs, err := b.Pick(context.Background(), troupe.Preferences{})
	if err != nil {
		t.Fatalf("Pick: %v", err)
	}
	if len(configs) != 1 {
		t.Errorf("Pick with empty prefs: got %d configs, want 1", len(configs))
	}
}

func TestDefaultBroker_Pick_ExplicitCount(t *testing.T) {
	b := broker.New("")
	configs, err := b.Pick(context.Background(), troupe.Preferences{Count: 3, Role: "worker"})
	if err != nil {
		t.Fatalf("Pick: %v", err)
	}
	if len(configs) != 3 {
		t.Errorf("Pick count=3: got %d configs, want 3", len(configs))
	}
}

func TestDefaultBroker_Spawn_NoLauncher(t *testing.T) {
	b := broker.New("")
	_, err := b.Spawn(context.Background(), troupe.ActorConfig{Model: "sonnet"})
	if err == nil {
		t.Fatal("expected error for spawn without launcher")
	}
}

// --- Multi-driver tests (TSK-10) ---

type providerDriver struct {
	started map[world.EntityID]bool
}

func newProviderDriver() *providerDriver {
	return &providerDriver{started: make(map[world.EntityID]bool)}
}

func (d *providerDriver) Start(_ context.Context, id world.EntityID, _ troupe.ActorConfig) error {
	d.started[id] = true
	return nil
}

func (d *providerDriver) Stop(_ context.Context, _ world.EntityID) error { return nil }

func TestBroker_MultiDriver_RoutesToProvider(t *testing.T) {
	anthropic := newProviderDriver()
	openai := newProviderDriver()
	b := broker.New("",
		broker.WithDriverFor("anthropic", anthropic),
		broker.WithDriverFor("openai", openai),
	)

	_, err := b.Spawn(context.Background(), troupe.ActorConfig{Provider: "anthropic", Role: "test"})
	if err != nil {
		t.Fatalf("Spawn anthropic: %v", err)
	}
	if len(anthropic.started) == 0 {
		t.Error("anthropic driver not called")
	}
	if len(openai.started) != 0 {
		t.Error("openai driver should not be called")
	}
}

func TestBroker_MultiDriver_FallbackToDefault(t *testing.T) {
	defaultD := newProviderDriver()
	b := broker.New("", broker.WithDriver(defaultD))

	_, err := b.Spawn(context.Background(), troupe.ActorConfig{Provider: "unknown", Role: "test"})
	if err != nil {
		t.Fatalf("Spawn with fallback: %v", err)
	}
	if len(defaultD.started) == 0 {
		t.Error("default driver not used as fallback")
	}
}

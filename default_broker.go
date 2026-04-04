package jericho

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/dpopsuev/jericho/identity"
	"github.com/dpopsuev/jericho/internal/acp"
	"github.com/dpopsuev/jericho/internal/agent"
	"github.com/dpopsuev/jericho/internal/transport"
	"github.com/dpopsuev/jericho/internal/warden"
	"github.com/dpopsuev/jericho/signal"
	"github.com/dpopsuev/jericho/world"
)

// ErrNoLauncher is returned when Spawn is called without a configured launcher.
var ErrNoLauncher = errors.New("broker: no launcher configured")

// DefaultBroker is the standard Broker implementation. Wires World, Warden,
// Transport, ACP, Registry, and Signal Bus internally. Staff is absorbed —
// DefaultBroker IS the agent subsystem orchestrator.
type DefaultBroker struct {
	world     *world.World
	warden    *warden.AgentWarden
	transport *transport.LocalTransport
	bus       signal.Bus
	registry  *identity.Registry
}

// NewBroker creates a Broker. If the endpoint is a remote URL (https://),
// returns a RemoteBroker that proxies over HTTP. Otherwise, returns a
// local DefaultBroker with ACP baked in.
func NewBroker(endpoint string) Broker {
	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		return newRemoteBroker(endpoint)
	}
	return newLocalBroker()
}

// newLocalBroker creates an in-process DefaultBroker with ACP.
func newLocalBroker() *DefaultBroker {
	w := world.NewWorld()
	t := transport.NewLocalTransport()
	b := signal.NewMemBus()
	p := warden.NewWarden(w, t, b, acp.NewACPLauncher())

	return &DefaultBroker{
		world:     w,
		warden:    p,
		transport: t,
		bus:       b,
		registry:  identity.NewRegistry(),
	}
}

// Pick returns actor configs matching preferences.
func (b *DefaultBroker) Pick(_ context.Context, prefs Preferences) ([]ActorConfig, error) {
	count := prefs.Count
	if count <= 0 {
		count = 1
	}

	configs := make([]ActorConfig, count)
	for i := range count {
		configs[i] = ActorConfig{
			Model: prefs.Model,
			Role:  prefs.Role,
		}
	}
	return configs, nil
}

// Spawn creates a running actor. Internally: Warden forks process, World
// creates entity, Transport registers handler.
func (b *DefaultBroker) Spawn(ctx context.Context, config ActorConfig) (Actor, error) {
	role := config.Role
	if role == "" {
		role = "actor"
	}

	id, err := b.warden.Fork(ctx, role, warden.AgentConfig{
		Model: config.Model,
	}, 0)
	if err != nil {
		return nil, fmt.Errorf("broker spawn: %w", err)
	}

	return agent.NewSolo(id, role, b.world, b.warden, b.transport), nil
}

// Signal returns the event bus.
func (b *DefaultBroker) Signal() signal.Bus { return b.bus }

// World returns the underlying ECS world (for advanced consumers).
func (b *DefaultBroker) World() *world.World { return b.world }

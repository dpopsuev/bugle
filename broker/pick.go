package broker

import (
	"context"

	troupe "github.com/dpopsuev/troupe"
)

// PickStrategy selects from candidate ActorConfigs before spawning.
// Pluggable: consumers can implement custom selection logic.
type PickStrategy interface {
	Choose(ctx context.Context, candidates []troupe.ActorConfig, prefs troupe.Preferences) []troupe.ActorConfig
}

// FirstMatch returns the first N candidates. Default strategy, backward compatible.
type FirstMatch struct{}

// Choose returns up to prefs.Count candidates from the front of the list.
func (FirstMatch) Choose(_ context.Context, candidates []troupe.ActorConfig, prefs troupe.Preferences) []troupe.ActorConfig {
	count := prefs.Count
	if count <= 0 {
		count = 1
	}
	if count > len(candidates) {
		count = len(candidates)
	}
	return candidates[:count]
}

// WithPickStrategy sets the actor selection strategy. Default: FirstMatch.
func WithPickStrategy(s PickStrategy) Option {
	return func(c *config) { c.pickStrategy = s }
}

// PickStrategyFrom wraps a Pick[ActorConfig] as a PickStrategy.
func PickStrategyFrom(p troupe.Pick[troupe.ActorConfig]) PickStrategy {
	return &pickAdapter{pick: p}
}

type pickAdapter struct {
	pick troupe.Pick[troupe.ActorConfig]
}

func (a *pickAdapter) Choose(ctx context.Context, candidates []troupe.ActorConfig, _ troupe.Preferences) []troupe.ActorConfig {
	return a.pick(ctx, candidates)
}

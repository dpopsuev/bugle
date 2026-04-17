# Claude Code Instructions for Troupe Development

## What is Troupe

Troupe is an AI agent broker — ECS framework for multi-agent orchestration without vendor lock-in. Three core interfaces (Broker, Actor, Director) compose agents, strategies, and drivers into collectives.

- Repo: github.com/dpopsuev/troupe
- Scribe scope: troupe (legacy artifacts may use BGL- or JRC- prefixes)

## Ecosystem Dependency Rules

**CRITICAL: Troupe is the bottom of the dependency stack.**

- Troupe NEVER imports origami/ or djinn/ or hegemony/
- Troupe defines interfaces (Actor, Broker, Director, Driver, Meter), consumers implement them
- Consumer-to-consumer communication goes through the protocol layer, not Go imports

Dependency direction: `Origami -> Troupe <- Djinn`

## Package Map

```
Root package   — Broker, Actor, Director, Driver, Meter interfaces
arsenal/       — Embedded model catalog (trait-scored selection, snapshot pinning)
billing/       — Token/cost tracking (CostBill, period management, enforcement)
broker/        — Broker implementation, multi-driver adapter, hooked actors
collective/    — Multi-agent collectives (Dialectic, Arbiter, pluggable Selector/Executor)
director/      — Director interface (orchestration contract)
execution/     — Provider config wrapper (ConfiguredProvider, any-llm integration)
identity/      — Agent archetypes (4 Thesis + 1 Antithesis), Color, Shade, Palette
referee/       — Event-driven scoring engine (YAML-defined weighted Scorecard rules)
resilience/    — Circuit breaker, retry, timeout (pure algorithms)
signal/        — Event bus (Bus, DurableBus), Andon health probes, tracing, EventStore
world/         — ECS entity-component store (Alive/Ready, ComponentType, hierarchy edges)
testkit/       — Test fixtures (MockActor, StubActor, LinearDirector, FanoutDirector)
```

### Internal packages

```
internal/acp/       — Agent Context Protocol launcher (JSON-RPC over stdio)
internal/agent/     — Solo agent implementation (Actor wrapper)
internal/auth/      — Authentication abstraction (Authenticate interface)
internal/protocol/  — Wire protocol types (message envelopes, contracts)
internal/transport/ — A2A messaging (LocalTransport, in-process channels)
internal/warden/    — Agent process supervision (Fork/Kill/Wait, restart, zombie reaping)
```

## Naming Conventions

- **Core interfaces**: Actor (Perform/Ready/Kill), Broker, Director, Driver, Meter
- **Health signals**: Andon (IEC 60073 stack light). NOT horn.
- **Events**: EventKind (Started, Completed, Failed, Transition, Done)
- **Archetypes**: Thesis (Sorter, Seeker, etc.) + Antithesis in identity/
- **Project name**: Troupe. NOT Jericho or Bugle.

## Go Conventions

- Go 1.25+
- golangci-lint enforced via pre-commit hook
- American English spelling (canceled, not cancelled)
- Sentinel errors with descriptive names
- slog for structured logging with constant key names

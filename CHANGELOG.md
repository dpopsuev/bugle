# Changelog

## v0.1.0 — Troupe (formerly Jericho)

### Breaking Changes

**Module rename**: `github.com/dpopsuev/jericho` → `github.com/dpopsuev/troupe`

Update all imports:
```go
// Before
import "github.com/dpopsuev/jericho"
// After
import "github.com/dpopsuev/troupe"
```

**Package rename**: `package jericho` → `package troupe`
- All references: `jericho.Broker` → `troupe.Broker`, etc.

**Protocol version**: `jericho/v1` → `troupe/v1`
- Settable via ldflag: `-ldflags "-X github.com/dpopsuev/troupe/internal/protocol.ProtocolVersion=custom/v1"`

**collective.Scale()**: `warden.AgentConfig` → `troupe.ActorConfig`
- Consumers no longer need to import `internal/warden`

**testkit**: `assert.go`, `handlers.go`, `quick.go` are now test-only
- `QuickWorld`, `QuickTransport`, `EchoHandler` no longer exported
- Use `MockBroker` and `MockActor` for consumer tests

**resilience/**: Promoted from `internal/resilience/` to public
- `CircuitBreaker`, `Retry`, `RateLimiter` now importable
- `internal/resilience/` retains only protocol-specific `responder.go`

### New Features

- **Hook interface** (`hook.go`): `SpawnHook`, `PerformHook` for lifecycle interception
- **DriverDescriptor** (`driver.go`): Optional capability declaration for Drivers
- **DriverValidator** (`driver.go`): Optional pre-flight environment checks
- **Multi-driver Broker**: `WithDriverFor(provider, driver)` routes by provider
- **PickStrategy** (`pick.go`): Pluggable actor selection. Built-in: `FirstMatch`
- **Meter** (`meter.go`): Provider-agnostic usage tracking. `InMemoryMeter` included
- **RetryActor** (`resilience/actor.go`): Decorator wrapping Actor with retry
- **FallbackActor** (`resilience/fallback.go`): Decorator with fallback chain
- **BudgetHook** (`billing/hook.go`): Pre-spawn budget enforcement via hooks
- **PeriodBudget** (`billing/period.go`): Time-period budget with lazy reset

### Consumer Migration

| Old Import | New Import |
|------------|------------|
| `github.com/dpopsuev/jericho` | `github.com/dpopsuev/troupe` |
| `jericho/identity` | `troupe/identity` |
| `jericho/collective` | `troupe/collective` |
| `jericho/signal` | `troupe/signal` |
| `jericho/world` | `troupe/world` |
| `jericho/billing` | `troupe/billing` |
| `jericho/arsenal` | `troupe/arsenal` |
| `jericho/internal/resilience` | `troupe/resilience` (public) |

### Deleted Packages
- `workload/` — dead code, 0 callers
- `orchestrate/` — absorbed into `internal/protocol/` + `testkit/mcp/`

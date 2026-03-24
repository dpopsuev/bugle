package guard

import (
	"context"
	"errors"
	"sync"
	"time"

	bd "github.com/dpopsuev/bugle/dispatch"
)

// CircuitState represents the current state of a circuit breaker.
type CircuitState int

const (
	CircuitClosed   CircuitState = iota // normal operation
	CircuitOpen                         // failures exceeded threshold, rejecting calls
	CircuitHalfOpen                     // cooldown elapsed, probing with one call
)

func (s CircuitState) String() string {
	switch s {
	case CircuitClosed:
		return "closed"
	case CircuitOpen:
		return "open"
	case CircuitHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// ErrCircuitOpen is returned when the circuit breaker is open and not yet
// in the cooldown-elapsed half-open state.
var ErrCircuitOpen = errors.New("circuit breaker is open")

// CircuitBreakerHook is called on circuit state transitions.
type CircuitBreakerHook func(from, to CircuitState)

// CircuitBreakerConfig configures a CircuitBreakerDispatcher.
type CircuitBreakerConfig struct {
	Threshold int           // consecutive failures before opening (default 5)
	Cooldown  time.Duration // duration to wait before half-open probe (default 30s)
	OnChange  CircuitBreakerHook
}

// CircuitBreakerDispatcher wraps a bd.Dispatcher with circuit breaker protection.
// After Threshold consecutive failures the circuit opens, rejecting all calls
// immediately with ErrCircuitOpen. After Cooldown elapses, one probe call is
// allowed (half-open): success closes, failure re-opens.
type CircuitBreakerDispatcher struct {
	inner     bd.Dispatcher
	threshold int
	cooldown  time.Duration
	onChange  CircuitBreakerHook

	mu       sync.Mutex
	state    CircuitState
	failures int
	openedAt time.Time
}

// NewCircuitBreakerDispatcher wraps inner with circuit breaker protection.
func NewCircuitBreakerDispatcher(inner bd.Dispatcher, cfg CircuitBreakerConfig) *CircuitBreakerDispatcher {
	threshold := cfg.Threshold
	if threshold <= 0 {
		threshold = 5
	}
	cooldown := cfg.Cooldown
	if cooldown <= 0 {
		cooldown = 30 * time.Second
	}
	return &CircuitBreakerDispatcher{
		inner:     inner,
		threshold: threshold,
		cooldown:  cooldown,
		onChange:  cfg.OnChange,
		state:     CircuitClosed,
	}
}

// State returns the current circuit state.
func (d *CircuitBreakerDispatcher) State() CircuitState {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.state
}

// Dispatch delegates to the inner dispatcher if the circuit is closed or
// half-open (probe). Returns ErrCircuitOpen if the circuit is open and
// cooldown has not elapsed.
func (d *CircuitBreakerDispatcher) Dispatch(ctx context.Context, dc bd.Context) ([]byte, error) { //nolint:gocritic // value receiver for API compat
	d.mu.Lock()
	if d.state == CircuitOpen {
		if time.Since(d.openedAt) >= d.cooldown {
			d.transition(CircuitHalfOpen)
		} else {
			d.mu.Unlock()
			return nil, ErrCircuitOpen
		}
	}
	d.mu.Unlock()

	data, err := d.inner.Dispatch(ctx, dc)

	d.mu.Lock()
	defer d.mu.Unlock()

	if err != nil {
		d.failures++
		if d.state == CircuitHalfOpen {
			d.transition(CircuitOpen)
			d.openedAt = time.Now()
		} else if d.failures >= d.threshold {
			d.transition(CircuitOpen)
			d.openedAt = time.Now()
		}
		return data, err
	}

	if d.state == CircuitHalfOpen || d.failures > 0 {
		d.transition(CircuitClosed)
	}
	d.failures = 0
	return data, nil
}

// Inner returns the wrapped dispatcher.
func (d *CircuitBreakerDispatcher) Inner() bd.Dispatcher { return d.inner }

func (d *CircuitBreakerDispatcher) transition(to CircuitState) {
	from := d.state
	if from == to {
		return
	}
	d.state = to
	if d.onChange != nil {
		d.onChange(from, to)
	}
}

package guard

import (
	"context"
	"sync/atomic"

	bd "github.com/dpopsuev/bugle/dispatch"

	"golang.org/x/time/rate"
)

// RateLimitHook is called each time a dispatch is delayed by the rate limiter.
type RateLimitHook func()

// RateLimitConfig configures a RateLimitDispatcher.
type RateLimitConfig struct {
	Rate    float64 // requests per second
	Burst   int     // max burst size (tokens available immediately)
	OnLimit RateLimitHook
}

// RateLimitDispatcher wraps a bd.Dispatcher with token bucket rate limiting.
// Dispatch calls block until a token is available, protecting downstream
// providers from burst traffic.
type RateLimitDispatcher struct {
	inner   bd.Dispatcher
	limiter *rate.Limiter
	onLimit RateLimitHook
	waits   atomic.Int64
}

// NewRateLimitDispatcher wraps inner with rate limiting.
func NewRateLimitDispatcher(inner bd.Dispatcher, cfg RateLimitConfig) *RateLimitDispatcher {
	r := cfg.Rate
	if r <= 0 {
		r = 10
	}
	burst := cfg.Burst
	if burst <= 0 {
		burst = 1
	}
	return &RateLimitDispatcher{
		inner:   inner,
		limiter: rate.NewLimiter(rate.Limit(r), burst),
		onLimit: cfg.OnLimit,
	}
}

// Dispatch waits for a rate limit token, then delegates to the inner dispatcher.
func (d *RateLimitDispatcher) Dispatch(ctx context.Context, dc bd.Context) ([]byte, error) { //nolint:gocritic // value receiver for API compat
	if !d.limiter.Allow() {
		d.waits.Add(1)
		if d.onLimit != nil {
			d.onLimit()
		}
		if err := d.limiter.Wait(ctx); err != nil {
			return nil, err
		}
	}
	return d.inner.Dispatch(ctx, dc)
}

// Waits returns the total number of times a dispatch was delayed.
func (d *RateLimitDispatcher) Waits() int64 { return d.waits.Load() }

// Inner returns the wrapped dispatcher.
func (d *RateLimitDispatcher) Inner() bd.Dispatcher { return d.inner }

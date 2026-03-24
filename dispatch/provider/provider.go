package provider

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	bd "github.com/dpopsuev/bugle/dispatch"
)

// Router selects a bd.Dispatcher based on the provider name carried
// in the bd.Context. This enables per-step LLM routing: one node uses
// Cursor (MuxDispatcher), another uses Codex (CLIDispatcher), a third
// calls OpenAI directly (HTTPDispatcher).
//
// If no provider is set in the context, StepProviderHints is checked for a
// fallback mapping (populated by Ouroboros PersonaSheet auto-routing).
// If neither is set, the Default dispatcher is used.
// If a provider is set but not found in the Routes map, an error is returned.
//
// When Fallbacks are configured for a provider and the primary dispatch fails,
// the router iterates through the fallback chain until one succeeds or all fail.
type Router struct {
	Default           bd.Dispatcher
	Routes            map[string]bd.Dispatcher
	StepProviderHints map[string]string   // step name → provider (populated by auto-routing)
	Fallbacks         map[string][]string // provider → ordered fallback provider names
	Logger            *slog.Logger
	OnFallback        func(primary, fallback string, err error) // optional callback on fallback activation
}

// RouterOption configures a Router.
type RouterOption func(*Router)

// NewRouter creates a router with a default dispatcher and optional routes.
func NewRouter(defaultDispatcher bd.Dispatcher, routes map[string]bd.Dispatcher, opts ...RouterOption) *Router {
	if routes == nil {
		routes = make(map[string]bd.Dispatcher)
	}
	r := &Router{
		Default: defaultDispatcher,
		Routes:  routes,
		Logger:  bd.DiscardLogger(),
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// WithProviderLogger sets a structured logger.
func WithProviderLogger(l *slog.Logger) RouterOption {
	return func(r *Router) { r.Logger = l }
}

// WithFallbacks configures fallback chains for providers.
func WithFallbacks(fallbacks map[string][]string) RouterOption {
	return func(r *Router) { r.Fallbacks = fallbacks }
}

// WithFallbackCallback sets a callback invoked when a fallback provider is used.
func WithFallbackCallback(fn func(primary, fallback string, err error)) RouterOption {
	return func(r *Router) { r.OnFallback = fn }
}

// Register adds a named provider route. Overwrites if the name already exists.
func (r *Router) Register(provider string, d bd.Dispatcher) {
	r.Routes[provider] = d
}

// Dispatch selects the appropriate dispatcher and delegates.
// On failure, iterates through the fallback chain if configured.
func (r *Router) Dispatch(ctx context.Context, dc bd.Context) ([]byte, error) { //nolint:gocritic // value receiver for API compat
	if dc.Provider == "" && r.StepProviderHints != nil {
		if hint, ok := r.StepProviderHints[dc.Step]; ok {
			if d, found := r.Routes[hint]; found {
				r.Logger.Debug("provider router: auto-route from PersonaSheet",
					slog.String("provider", hint),
					slog.String("step", dc.Step),
				)
				return r.dispatchWithFallback(ctx, hint, d, dc)
			}
		}
	}

	if dc.Provider == "" {
		r.Logger.Debug("provider router: using default dispatcher",
			slog.String("case_id", dc.CaseID),
			slog.String("step", dc.Step),
		)
		return r.dispatchWithFallback(ctx, "default", r.Default, dc)
	}

	d, ok := r.Routes[dc.Provider]
	if !ok {
		return nil, fmt.Errorf("dispatch/provider: unknown provider %q (registered: %v)",
			dc.Provider, r.providerNames())
	}

	r.Logger.Debug("provider router: routing to provider",
		slog.String("provider", dc.Provider),
		slog.String("case_id", dc.CaseID),
		slog.String("step", dc.Step),
	)
	return r.dispatchWithFallback(ctx, dc.Provider, d, dc)
}

// dispatchWithFallback tries the primary dispatcher, then iterates through
// fallbacks on failure. Returns the first successful result or an aggregated error.
func (r *Router) dispatchWithFallback(ctx context.Context, providerName string, primary bd.Dispatcher, dc bd.Context) ([]byte, error) { //nolint:gocritic // value receiver for API compat
	result, err := primary.Dispatch(ctx, dc)
	if err == nil {
		return result, nil
	}

	chain := r.Fallbacks[providerName]
	if len(chain) == 0 {
		return nil, err
	}

	errs := make([]error, 0, 1+len(chain))
	errs = append(errs, fmt.Errorf("primary %s: %w", providerName, err))

	for _, fb := range chain {
		d, ok := r.Routes[fb]
		if !ok {
			errs = append(errs, fmt.Errorf("fallback %s: not registered", fb))
			continue
		}

		r.Logger.Info("provider router: fallback activated",
			slog.String("primary", providerName),
			slog.String("fallback", fb),
			slog.String("step", dc.Step),
		)
		if r.OnFallback != nil {
			r.OnFallback(providerName, fb, err)
		}

		result, fbErr := d.Dispatch(ctx, dc)
		if fbErr == nil {
			return result, nil
		}
		errs = append(errs, fmt.Errorf("fallback %s: %w", fb, fbErr))
	}

	return nil, fmt.Errorf("all providers failed: %w", errors.Join(errs...))
}

func (r *Router) providerNames() []string {
	names := make([]string, 0, len(r.Routes))
	for k := range r.Routes {
		names = append(names, k)
	}
	return names
}

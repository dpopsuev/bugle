package dispatch

import (
	"fmt"
	"net/http"
	"os"
)

// StaticTokenAuth is an http.RoundTripper middleware that injects a bearer
// token from an environment variable into every outgoing request.
//
// This is a PoC battery — sufficient for prototyping, not production-grade.
// Consumers should replace it with their own auth for production use.
type StaticTokenAuth struct {
	EnvVar string
	Inner  http.RoundTripper
}

// NewStaticTokenAuth creates a RoundTripper that reads the bearer token from
// the given environment variable and injects it into the Authorization header.
// If inner is nil, http.DefaultTransport is used.
func NewStaticTokenAuth(envVar string, inner http.RoundTripper) *StaticTokenAuth {
	if inner == nil {
		inner = http.DefaultTransport
	}
	return &StaticTokenAuth{EnvVar: envVar, Inner: inner}
}

func (a *StaticTokenAuth) RoundTrip(req *http.Request) (*http.Response, error) {
	token := os.Getenv(a.EnvVar)
	if token == "" {
		return nil, fmt.Errorf("dispatch/auth: %s environment variable not set", a.EnvVar)
	}
	clone := req.Clone(req.Context())
	clone.Header.Set("Authorization", "Bearer "+token)
	return a.Inner.RoundTrip(clone)
}

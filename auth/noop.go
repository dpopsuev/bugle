// Package auth provides Authenticator and Authorizer adapters for the
// Bugle Protocol. Consumers select the adapter matching their infrastructure.
package auth

import (
	"context"

	"github.com/dpopsuev/jericho/work"
)

// Noop allows all requests. Use for development and testing.
type Noop struct{}

// Authenticate always succeeds with a generic symbol.
func (Noop) Authenticate(_ context.Context, token string) (work.Identity, error) {
	return work.Identity{Subject: "anonymous"}, nil
}

// Authorize always allows.
func (Noop) Authorize(_ work.Identity, _ work.Action) error {
	return nil
}

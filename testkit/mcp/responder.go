// Package testkit provides test doubles for the Bugle Protocol.
// Mock responders, mock servers, assertion helpers, and conformance contracts.
package mcp

import (
	"context"
	"fmt"
	"sync"
)

// StaticResponder always returns the same response.
type StaticResponder struct {
	Response string
}

// RespondTo returns the static response.
func (r *StaticResponder) RespondTo(_ context.Context, _ string) (string, error) {
	return r.Response, nil
}

// FailingResponder always returns an error.
type FailingResponder struct {
	Err error
}

// RespondTo returns the configured error.
func (r *FailingResponder) RespondTo(_ context.Context, _ string) (string, error) {
	return "", r.Err
}

// ScriptedResponder replays responses in sequence. Returns an error
// when the script is exhausted.
type ScriptedResponder struct {
	mu        sync.Mutex
	responses []string
	idx       int
}

// NewScriptedResponder creates a responder that replays the given responses in order.
func NewScriptedResponder(responses ...string) *ScriptedResponder {
	return &ScriptedResponder{responses: responses}
}

// RespondTo returns the next response in the script.
func (r *ScriptedResponder) RespondTo(_ context.Context, _ string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.idx >= len(r.responses) {
		return "", fmt.Errorf("scripted responder: exhausted after %d responses", len(r.responses))
	}
	resp := r.responses[r.idx]
	r.idx++
	return resp, nil
}

// CallCount returns how many times RespondTo was called.
func (r *ScriptedResponder) CallCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.idx
}

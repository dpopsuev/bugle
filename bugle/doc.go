// Package bugle defines the Bugle Protocol — a wire format for distributing
// work to autonomous agent workers over MCP.
//
// This package contains pure types with zero external dependencies.
// Action constants, request/response structs, horn levels, status codes,
// budget types, capability negotiation, and error codes.
//
// The protocol has five layers:
//   - Core: start, step, submit, cancel
//   - Health: horn severity signals (green/yellow/red/black)
//   - Budget: resource tracking (tokens, time, cost)
//   - HITL: human-in-the-loop (blocked/resolved, cordon/uncordon)
//   - Observability: aggregated status
//
// Spec: BGL-SPC-15
package bugle

package mcp

import (
	"context"
	"sync"

	"github.com/dpopsuev/jericho/work"
)

// MockServer implements work.Server with configurable handlers and call tracking.
type MockServer struct {
	mu sync.Mutex

	startFn  func(work.StartRequest) (work.StartResponse, error)
	pullFn   func(work.PullRequest) (work.PullResponse, error)
	pushFn   func(work.PushRequest) (work.PushResponse, error)
	cancelFn func(work.CancelRequest) (work.CancelResponse, error)
	statusFn func(work.StatusRequest) (work.StatusResponse, error)

	pushes  []work.PushRequest
	pulls   int
	starts  int
	cancels int
}

// NewMockServer creates a mock with default handlers that return zero values.
func NewMockServer() *MockServer {
	return &MockServer{}
}

// OnStart sets the handler for start requests.
func (s *MockServer) OnStart(fn func(work.StartRequest) (work.StartResponse, error)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.startFn = fn
}

// OnPull sets the handler for pull requests.
func (s *MockServer) OnPull(fn func(work.PullRequest) (work.PullResponse, error)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pullFn = fn
}

// OnPush sets the handler for push requests.
func (s *MockServer) OnPush(fn func(work.PushRequest) (work.PushResponse, error)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pushFn = fn
}

// OnCancel sets the handler for cancel requests.
func (s *MockServer) OnCancel(fn func(work.CancelRequest) (work.CancelResponse, error)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cancelFn = fn
}

// OnStatus sets the handler for status requests.
func (s *MockServer) OnStatus(fn func(work.StatusRequest) (work.StatusResponse, error)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.statusFn = fn
}

// Start implements work.Server.
func (s *MockServer) Start(_ context.Context, req work.StartRequest) (work.StartResponse, error) {
	s.mu.Lock()
	s.starts++
	fn := s.startFn
	s.mu.Unlock()
	if fn != nil {
		return fn(req)
	}
	return work.StartResponse{SessionID: "mock-session", TotalItems: 0, Status: "running"}, nil
}

// Pull implements work.Server.
func (s *MockServer) Pull(_ context.Context, req work.PullRequest) (work.PullResponse, error) {
	s.mu.Lock()
	s.pulls++
	fn := s.pullFn
	s.mu.Unlock()
	if fn != nil {
		return fn(req)
	}
	return work.PullResponse{Done: true}, nil
}

// Push implements work.Server.
func (s *MockServer) Push(_ context.Context, req work.PushRequest) (work.PushResponse, error) {
	s.mu.Lock()
	s.pushes = append(s.pushes, req)
	fn := s.pushFn
	s.mu.Unlock()
	if fn != nil {
		return fn(req)
	}
	return work.PushResponse{OK: true}, nil
}

// Cancel implements work.Server.
func (s *MockServer) Cancel(_ context.Context, req work.CancelRequest) (work.CancelResponse, error) {
	s.mu.Lock()
	s.cancels++
	fn := s.cancelFn
	s.mu.Unlock()
	if fn != nil {
		return fn(req)
	}
	return work.CancelResponse{OK: true, Canceled: 1}, nil
}

// Status implements work.Server.
func (s *MockServer) Status(_ context.Context, req work.StatusRequest) (work.StatusResponse, error) {
	s.mu.Lock()
	fn := s.statusFn
	s.mu.Unlock()
	if fn != nil {
		return fn(req)
	}
	return work.StatusResponse{SessionID: req.SessionID, Progress: work.Progress{}}, nil
}

// --- Inspection methods ---

// PushCount returns the number of push calls received.
func (s *MockServer) PushCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.pushes)
}

// Pushes returns all push requests received.
func (s *MockServer) Pushes() []work.PushRequest {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := make([]work.PushRequest, len(s.pushes))
	copy(cp, s.pushes)
	return cp
}

// LastPush returns the most recent push request. Panics if none.
func (s *MockServer) LastPush() work.PushRequest {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.pushes[len(s.pushes)-1]
}

// PullCount returns the number of pull calls received.
func (s *MockServer) PullCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.pulls
}

// StartCount returns the number of start calls received.
func (s *MockServer) StartCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.starts
}

// CancelCount returns the number of cancel calls received.
func (s *MockServer) CancelCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cancels
}

package world

import (
	"errors"
	"testing"
)

func TestEdgeStore_LinkAndNeighbors(t *testing.T) {
	s := NewEdgeStore()

	if err := s.Link(1, Supervises, 2); err != nil {
		t.Fatal(err)
	}
	if err := s.Link(1, Supervises, 3); err != nil {
		t.Fatal(err)
	}

	// Outbound from 1.
	out := s.Neighbors(1, Supervises, Outbound)
	if len(out) != 2 {
		t.Fatalf("outbound = %d, want 2", len(out))
	}

	// Inbound to 2.
	in := s.Neighbors(2, Supervises, Inbound)
	if len(in) != 1 || in[0] != 1 {
		t.Fatalf("inbound to 2 = %v, want [1]", in)
	}

	// Both directions from 1.
	both := s.Neighbors(1, Supervises, Both)
	if len(both) != 2 {
		t.Fatalf("both = %d, want 2", len(both))
	}
}

func TestEdgeStore_SelfLoopRejected(t *testing.T) {
	s := NewEdgeStore()
	err := s.Link(1, Supervises, 1)
	if !errors.Is(err, ErrSelfLoop) {
		t.Fatalf("expected ErrSelfLoop, got %v", err)
	}
}

func TestEdgeStore_DuplicateRejected(t *testing.T) {
	s := NewEdgeStore()
	_ = s.Link(1, Supervises, 2)
	err := s.Link(1, Supervises, 2)
	if !errors.Is(err, ErrDuplicateEdge) {
		t.Fatalf("expected ErrDuplicateEdge, got %v", err)
	}
}

func TestEdgeStore_CycleRejected(t *testing.T) {
	s := NewEdgeStore()
	_ = s.Link(1, Supervises, 2)
	_ = s.Link(2, Supervises, 3)

	// 3 → 1 would create cycle: 1→2→3→1.
	err := s.Link(3, Supervises, 1)
	if !errors.Is(err, ErrCycleDetected) {
		t.Fatalf("expected ErrCycleDetected, got %v", err)
	}
}

func TestEdgeStore_CycleNotEnforcedForNonDAG(t *testing.T) {
	s := NewEdgeStore()
	_ = s.Link(1, CommunicatesWith, 2)
	_ = s.Link(2, CommunicatesWith, 3)

	// Cycles allowed for communicates_with (not DAG-constrained).
	err := s.Link(3, CommunicatesWith, 1)
	if err != nil {
		t.Fatalf("communicates_with should allow cycles, got %v", err)
	}
}

func TestEdgeStore_Unlink(t *testing.T) {
	s := NewEdgeStore()
	_ = s.Link(1, Supervises, 2)
	_ = s.Link(1, Supervises, 3)

	if err := s.Unlink(1, Supervises, 2); err != nil {
		t.Fatal(err)
	}

	out := s.Neighbors(1, Supervises, Outbound)
	if len(out) != 1 || out[0] != 3 {
		t.Fatalf("after unlink = %v, want [3]", out)
	}

	if s.Count() != 1 {
		t.Fatalf("count = %d, want 1", s.Count())
	}
}

func TestEdgeStore_UnlinkNotFound(t *testing.T) {
	s := NewEdgeStore()
	err := s.Unlink(1, Supervises, 2)
	if !errors.Is(err, ErrEdgeNotFound) {
		t.Fatalf("expected ErrEdgeNotFound, got %v", err)
	}
}

func TestEdgeStore_Edges(t *testing.T) {
	s := NewEdgeStore()
	_ = s.Link(1, Supervises, 2)
	_ = s.Link(3, CommunicatesWith, 1)

	edges := s.Edges(1)
	if len(edges) != 2 {
		t.Fatalf("edges = %d, want 2 (1 out + 1 in)", len(edges))
	}
}

func TestEdgeStore_RemoveEntity(t *testing.T) {
	s := NewEdgeStore()
	_ = s.Link(1, Supervises, 2)
	_ = s.Link(1, Supervises, 3)
	_ = s.Link(3, CommunicatesWith, 1)

	s.RemoveEntity(1)

	if s.Count() != 0 {
		t.Fatalf("count after remove = %d, want 0", s.Count())
	}
	if len(s.Neighbors(2, Supervises, Inbound)) != 0 {
		t.Error("2 should have no inbound after 1 removed")
	}
}

// --- World-level edge API tests (TSK-84) ---

func TestWorld_LinkAndNeighbors(t *testing.T) {
	w := NewWorld()
	a := w.Spawn()
	b := w.Spawn()
	c := w.Spawn()

	if err := w.Link(a, Supervises, b); err != nil {
		t.Fatal(err)
	}
	if err := w.Link(a, Supervises, c); err != nil {
		t.Fatal(err)
	}

	children := w.Neighbors(a, Supervises, Outbound)
	if len(children) != 2 {
		t.Fatalf("children = %d, want 2", len(children))
	}
	if w.EdgeCount() != 2 {
		t.Fatalf("EdgeCount = %d, want 2", w.EdgeCount())
	}
}

func TestWorld_DespawnCleansEdges(t *testing.T) {
	w := NewWorld()
	a := w.Spawn()
	b := w.Spawn()
	c := w.Spawn()

	_ = w.Link(a, Supervises, b)
	_ = w.Link(a, Supervises, c)
	_ = w.Link(c, CommunicatesWith, a)

	if w.EdgeCount() != 3 {
		t.Fatalf("before despawn: EdgeCount = %d, want 3", w.EdgeCount())
	}

	w.Despawn(a)

	if w.EdgeCount() != 0 {
		t.Fatalf("after despawn: EdgeCount = %d, want 0", w.EdgeCount())
	}
}

func TestEdgeStore_DifferentRelationsSameEndpoints(t *testing.T) {
	s := NewEdgeStore()
	_ = s.Link(1, Supervises, 2)
	err := s.Link(1, CommunicatesWith, 2)
	if err != nil {
		t.Fatalf("different relation same endpoints should be allowed: %v", err)
	}
	if s.Count() != 2 {
		t.Fatalf("count = %d, want 2", s.Count())
	}
}

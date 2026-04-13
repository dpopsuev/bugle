// edge.go — typed directed edges between entities in the World ECS.
//
// Edges model relationships: supervises, assigned_to, communicates_with,
// member_of, flows_to. Stored in an adjacency list. Thread-safe.
//
// TRP-TSK-82, TRP-GOL-14
package world

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
)

// Sentinel errors.
var (
	ErrSelfLoop       = errors.New("edge: self-loop not allowed")
	ErrDuplicateEdge  = errors.New("edge: duplicate edge")
	ErrEdgeNotFound   = errors.New("edge: not found")
	ErrCycleDetected  = errors.New("edge: cycle detected (DAG constraint)")
)

// Relation is a typed edge label.
type Relation string

const (
	Supervises       Relation = "supervises"        // ownership, DAG-constrained
	AssignedTo       Relation = "assigned_to"        // work dispatch
	CommunicatesWith Relation = "communicates_with"  // scoped messaging
	MemberOf         Relation = "member_of"          // collective membership
	FlowsTo          Relation = "flows_to"           // circuit edge (Origami seam)
)

// dagRelations are constrained to be acyclic.
var dagRelations = map[Relation]bool{
	Supervises: true,
}

// Direction for neighbor queries.
type Direction int

const (
	Outbound Direction = iota
	Inbound
	Both
)

// Edge is a typed directed connection between two entities.
type Edge struct {
	From     EntityID `json:"from"`
	Relation Relation `json:"relation"`
	To       EntityID `json:"to"`
}

// edgeKey is the dedup key for edge storage.
type edgeKey struct {
	from     EntityID
	relation Relation
	to       EntityID
}

// EdgeStore manages typed directed edges between entities.
// Thread-safe via mutex.
type EdgeStore struct {
	mu       sync.RWMutex
	outbound map[EntityID][]Edge // from → edges
	inbound  map[EntityID][]Edge // to → edges
	keys     map[edgeKey]bool    // dedup
	log      *slog.Logger       // optional: ORANGE/YELLOW instrumentation
}

// NewEdgeStore creates an empty edge store.
func NewEdgeStore() *EdgeStore {
	return &EdgeStore{
		outbound: make(map[EntityID][]Edge),
		inbound:  make(map[EntityID][]Edge),
		keys:     make(map[edgeKey]bool),
	}
}

// WithLogger sets structured logging for edge operations.
func (s *EdgeStore) WithLogger(l *slog.Logger) *EdgeStore {
	s.log = l
	return s
}

// Link creates a directed edge. Returns error on self-loop, duplicate, or cycle.
func (s *EdgeStore) Link(from EntityID, rel Relation, to EntityID) error {
	if from == to {
		s.warnEdge("self-loop rejected", from, rel, to)
		return fmt.Errorf("%w: %d -[%s]-> %d", ErrSelfLoop, from, rel, to)
	}

	key := edgeKey{from, rel, to}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.keys[key] {
		return fmt.Errorf("%w: %d -[%s]-> %d", ErrDuplicateEdge, from, rel, to)
	}

	if dagRelations[rel] && s.reachesLocked(to, from, rel) {
		s.warnEdge("cycle rejected", from, rel, to)
		return fmt.Errorf("%w: %d -[%s]-> %d", ErrCycleDetected, from, rel, to)
	}

	edge := Edge{From: from, Relation: rel, To: to}
	s.outbound[from] = append(s.outbound[from], edge)
	s.inbound[to] = append(s.inbound[to], edge)
	s.keys[key] = true

	s.infoEdge("edge created", from, rel, to)
	return nil
}

func (s *EdgeStore) warnEdge(msg string, from EntityID, rel Relation, to EntityID) {
	if s.log != nil {
		s.log.WarnContext(context.Background(), msg,
			slog.Uint64("from", uint64(from)),
			slog.String("relation", string(rel)),
			slog.Uint64("to", uint64(to)),
		)
	}
}

func (s *EdgeStore) infoEdge(msg string, from EntityID, rel Relation, to EntityID) {
	if s.log != nil {
		s.log.DebugContext(context.Background(), msg,
			slog.Uint64("from", uint64(from)),
			slog.String("relation", string(rel)),
			slog.Uint64("to", uint64(to)),
			slog.Int("total_edges", len(s.keys)),
		)
	}
}

// Unlink removes a directed edge. Returns error if not found.
func (s *EdgeStore) Unlink(from EntityID, rel Relation, to EntityID) error {
	key := edgeKey{from, rel, to}

	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.keys[key] {
		return fmt.Errorf("%w: %d -[%s]-> %d", ErrEdgeNotFound, from, rel, to)
	}

	delete(s.keys, key)
	s.outbound[from] = removeEdge(s.outbound[from], rel, to)
	s.inbound[to] = removeEdge(s.inbound[to], rel, from)
	return nil
}

// Neighbors returns entity IDs connected to id by the given relation and direction.
func (s *EdgeStore) Neighbors(id EntityID, rel Relation, dir Direction) []EntityID {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var ids []EntityID
	if dir == Outbound || dir == Both {
		for _, e := range s.outbound[id] {
			if e.Relation == rel {
				ids = append(ids, e.To)
			}
		}
	}
	if dir == Inbound || dir == Both {
		for _, e := range s.inbound[id] {
			if e.Relation == rel {
				ids = append(ids, e.From)
			}
		}
	}
	return ids
}

// Edges returns all edges connected to id (both directions).
func (s *EdgeStore) Edges(id EntityID) []Edge {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var edges []Edge
	edges = append(edges, s.outbound[id]...)
	edges = append(edges, s.inbound[id]...)
	return edges
}

// Count returns total number of edges.
func (s *EdgeStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.keys)
}

// RemoveEntity removes all edges involving the entity.
func (s *EdgeStore) RemoveEntity(id EntityID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove outbound edges.
	for _, e := range s.outbound[id] {
		key := edgeKey{e.From, e.Relation, e.To}
		delete(s.keys, key)
		s.inbound[e.To] = removeEdge(s.inbound[e.To], e.Relation, e.From)
	}
	delete(s.outbound, id)

	// Remove inbound edges.
	for _, e := range s.inbound[id] {
		key := edgeKey{e.From, e.Relation, e.To}
		delete(s.keys, key)
		s.outbound[e.From] = removeEdge(s.outbound[e.From], e.Relation, e.To)
	}
	delete(s.inbound, id)
}

// reachesLocked checks if `from` can reach `target` via edges of the given relation.
// BFS. Caller must hold at least read lock.
func (s *EdgeStore) reachesLocked(from, target EntityID, rel Relation) bool {
	visited := map[EntityID]bool{from: true}
	queue := []EntityID{from}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		for _, e := range s.outbound[curr] {
			if e.Relation != rel {
				continue
			}
			if e.To == target {
				return true
			}
			if !visited[e.To] {
				visited[e.To] = true
				queue = append(queue, e.To)
			}
		}
	}
	return false
}

// removeEdge filters out one edge by relation + target from a slice.
func removeEdge(edges []Edge, rel Relation, target EntityID) []Edge {
	n := 0
	for _, e := range edges {
		// For outbound: target matches To. For inbound: target matches From.
		if e.Relation == rel && (e.To == target || e.From == target) {
			continue
		}
		edges[n] = e
		n++
	}
	return edges[:n]
}

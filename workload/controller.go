package workload

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dpopsuev/jericho/agent"
	"github.com/dpopsuev/jericho/warden"
)

// Slog key constants.
const (
	logKeyKind    = "kind"
	logKeyName    = "name"
	logKeySpawned = "spawned"
	logKeyKilled  = "killed"
	logKeyAgents  = "agents"
	logKeyWorkers = "workers"
)

// Controller reconciles desired workload state against actual agent state.
type Controller struct {
	staff *agent.Staff
}

// NewController creates a workload controller backed by the given Staff.
func NewController(staff *agent.Staff) *Controller {
	return &Controller{staff: staff}
}

// ApplyWorkerPool reconciles a WorkerPool — spawns or kills agents to match replicas.
func (c *Controller) ApplyWorkerPool(ctx context.Context, wp *WorkerPool) error {
	role := wp.Metadata.Name
	current := c.staff.FindByRole(role)
	desired := wp.Spec.Replicas
	delta := desired - len(current)

	if delta > 0 {
		for range delta {
			if _, err := c.staff.Spawn(ctx, role, wp.Spec.Template); err != nil {
				return fmt.Errorf("workload %s: spawn: %w", role, err)
			}
		}
		slog.InfoContext(ctx, "workload reconciled", logKeyKind, "WorkerPool", logKeyName, role, logKeySpawned, delta)
	} else if delta < 0 {
		for i := range -delta {
			if i < len(current) {
				current[i].Kill(ctx) //nolint:errcheck // best-effort during scale-down
			}
		}
		slog.InfoContext(ctx, "workload reconciled", logKeyKind, "WorkerPool", logKeyName, role, logKeyKilled, -delta)
	}

	return nil
}

// ApplyDebateTeam reconciles a DebateTeam — ensures all agents exist.
func (c *Controller) ApplyDebateTeam(ctx context.Context, dt *DebateTeam) error {
	role := dt.Metadata.Name
	current := c.staff.FindByRole(role)

	if len(current) >= len(dt.Spec.Agents) {
		return nil // already at desired count
	}

	for i := len(current); i < len(dt.Spec.Agents); i++ {
		cfg := dt.Spec.Agents[i]
		if _, err := c.staff.Spawn(ctx, role, cfg); err != nil {
			return fmt.Errorf("workload %s: spawn agent %d: %w", role, i, err)
		}
	}
	slog.InfoContext(ctx, "workload reconciled", logKeyKind, "DebateTeam", logKeyName, role, logKeyAgents, len(dt.Spec.Agents))
	return nil
}

// ApplyTaskRunner reconciles a TaskRunner — spawns workers for the session.
func (c *Controller) ApplyTaskRunner(ctx context.Context, tr *TaskRunner) error {
	role := tr.Metadata.Name
	current := c.staff.FindByRole(role)
	desired := tr.Spec.Workers

	for range desired - len(current) {
		cfg := tr.Spec.Template
		cfg.Role = role
		if _, err := c.staff.Spawn(ctx, role, cfg); err != nil {
			return fmt.Errorf("workload %s: spawn: %w", role, err)
		}
	}
	slog.InfoContext(ctx, "workload reconciled", logKeyKind, "TaskRunner", logKeyName, role, logKeyWorkers, desired)
	return nil
}

// Apply parses and applies any workload definition.
func (c *Controller) Apply(ctx context.Context, data []byte) error {
	w, err := Parse(data)
	if err != nil {
		return err
	}
	switch v := w.(type) {
	case *WorkerPool:
		return c.ApplyWorkerPool(ctx, v)
	case *DebateTeam:
		return c.ApplyDebateTeam(ctx, v)
	case *TaskRunner:
		return c.ApplyTaskRunner(ctx, v)
	default:
		return fmt.Errorf("%w: %T", ErrUnsupportedKind, v)
	}
}

// FindByRole is needed from Staff — check it exists.
var _ interface {
	FindByRole(role string) []*agent.Solo
} = (*agent.Staff)(nil)

// Ensure warden.AgentConfig has YAML tags (it does via Phase 3).
var _ = warden.AgentConfig{}

// Package workload defines declarative agent workload definitions.
// K8s-inspired YAML specs for agent groups: WorkerPool (Deployment),
// DebateTeam (StatefulSet), TaskRunner (Job).
package workload

import (
	"errors"

	"github.com/dpopsuev/jericho/pool"
)

// Sentinel errors.
var (
	ErrUnknownKind     = errors.New("workload: unknown kind")
	ErrUnsupportedKind = errors.New("workload: unsupported kind")
)

// Kind identifies the workload type.
type Kind string

const (
	KindWorkerPool Kind = "WorkerPool"
	KindDebateTeam Kind = "DebateTeam"
	KindTaskRunner Kind = "TaskRunner"
)

// Metadata is the standard metadata block for all workload types.
type Metadata struct {
	Name   string            `yaml:"name"`
	Labels map[string]string `yaml:"labels,omitempty"`
}

// WorkerPool is a Deployment analog — N identical stateless agents.
type WorkerPool struct {
	Kind     Kind           `yaml:"kind"`
	Metadata Metadata       `yaml:"metadata"`
	Spec     WorkerPoolSpec `yaml:"spec"`
}

// WorkerPoolSpec defines the desired state of a WorkerPool.
type WorkerPoolSpec struct {
	Replicas int              `yaml:"replicas"`
	Template pool.AgentConfig `yaml:"template"`
}

// DebateTeam is a StatefulSet analog — agents with ordinal identity
// and a collective strategy.
type DebateTeam struct {
	Kind     Kind           `yaml:"kind"`
	Metadata Metadata       `yaml:"metadata"`
	Spec     DebateTeamSpec `yaml:"spec"`
}

// DebateTeamSpec defines the desired state of a DebateTeam.
type DebateTeamSpec struct {
	Agents   []pool.AgentConfig `yaml:"agents"`
	Strategy string             `yaml:"strategy"` // dialectic, arbiter, dialectic-pair
	Shade    string             `yaml:"shade,omitempty"`
}

// TaskRunner is a Job analog — run workers until completions reached.
type TaskRunner struct {
	Kind     Kind           `yaml:"kind"`
	Metadata Metadata       `yaml:"metadata"`
	Spec     TaskRunnerSpec `yaml:"spec"`
}

// TaskRunnerSpec defines the desired state of a TaskRunner.
type TaskRunnerSpec struct {
	Session      string           `yaml:"session"`
	Endpoint     string           `yaml:"endpoint"`
	Workers      int              `yaml:"workers"`
	Completions  int              `yaml:"completions"`
	BackoffLimit int              `yaml:"backoff_limit,omitempty"`
	Template     pool.AgentConfig `yaml:"template,omitempty"`
}

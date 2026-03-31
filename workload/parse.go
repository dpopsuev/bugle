package workload

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// kindHeader is used for initial YAML parsing to determine the workload kind.
type kindHeader struct {
	Kind Kind `yaml:"kind"`
}

// Parse reads YAML and returns the typed workload definition.
func Parse(data []byte) (any, error) {
	var header kindHeader
	if err := yaml.Unmarshal(data, &header); err != nil {
		return nil, fmt.Errorf("workload: parse kind: %w", err)
	}

	switch header.Kind {
	case KindWorkerPool:
		var w WorkerPool
		if err := yaml.Unmarshal(data, &w); err != nil {
			return nil, fmt.Errorf("workload: parse WorkerPool: %w", err)
		}
		return &w, nil

	case KindDebateTeam:
		var d DebateTeam
		if err := yaml.Unmarshal(data, &d); err != nil {
			return nil, fmt.Errorf("workload: parse DebateTeam: %w", err)
		}
		return &d, nil

	case KindTaskRunner:
		var t TaskRunner
		if err := yaml.Unmarshal(data, &t); err != nil {
			return nil, fmt.Errorf("workload: parse TaskRunner: %w", err)
		}
		return &t, nil

	default:
		return nil, fmt.Errorf("%w: %q", ErrUnknownKind, header.Kind)
	}
}

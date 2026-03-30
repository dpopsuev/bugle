package dispatch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ErrNoArtifact is returned when no pre-authored artifact exists for a case+step pair.
var ErrNoArtifact = errors.New("static dispatcher: no artifact")

// StaticDispatcher returns pre-authored artifact data by looking up
// (CaseID, Step) in a directory of JSON files. Used for deterministic
// calibration without LLM variance.
//
// Directory layout:
//
//	artifacts_dir/
//	  C1/
//	    F0_RECALL.json
//	    F1_TRIAGE.json
//	  C2/
//	    F0_RECALL.json
//
// Alternatively, artifacts can be registered in-memory via Set().
type StaticDispatcher struct {
	dir       string
	artifacts map[string]json.RawMessage // "case_id:step" → raw JSON
}

// NewStaticDispatcher creates a dispatcher that returns pre-authored artifacts.
// If dir is non-empty, artifacts are loaded from files at dispatch time.
func NewStaticDispatcher(dir string) *StaticDispatcher {
	return &StaticDispatcher{
		dir:       dir,
		artifacts: make(map[string]json.RawMessage),
	}
}

// Set registers an in-memory artifact for a case+step pair.
func (d *StaticDispatcher) Set(caseID, step string, data json.RawMessage) {
	d.artifacts[staticKey(caseID, step)] = data
}

// Dispatch returns the pre-authored artifact for the given case and step.
func (d *StaticDispatcher) Dispatch(_ context.Context, ctx Context) ([]byte, error) { //nolint:gocritic // value receiver for API compat
	key := staticKey(ctx.CaseID, ctx.Step)

	if data, ok := d.artifacts[key]; ok {
		return data, nil
	}

	if d.dir != "" {
		path := filepath.Join(d.dir, ctx.CaseID, ctx.Step+".json")
		data, err := os.ReadFile(path)
		if err == nil {
			return data, nil
		}
		path = filepath.Join(d.dir, ctx.CaseID, strings.ToLower(ctx.Step)+".json")
		data, err = os.ReadFile(path)
		if err == nil {
			return data, nil
		}
	}

	return nil, fmt.Errorf("%w for %s/%s", ErrNoArtifact, ctx.CaseID, ctx.Step)
}

func staticKey(caseID, step string) string {
	return caseID + ":" + step
}

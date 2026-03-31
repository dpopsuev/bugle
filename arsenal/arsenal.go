package arsenal

import (
	"fmt"
)

// Arsenal combines catalog snapshots with consumer preferences for model selection.
type Arsenal struct {
	snapshots map[string]*Snapshot
	active    string // resolved pin
}

// NewArsenal loads the embedded catalog and resolves the pin.
// Pin "" or "latest" uses the newest snapshot. Explicit pin (e.g. "2026-03")
// selects that snapshot.
func NewArsenal(pin string) (*Arsenal, error) {
	names, err := availableSnapshots()
	if err != nil {
		return nil, err
	}
	if len(names) == 0 {
		return nil, ErrEmptyCatalog
	}

	a := &Arsenal{snapshots: make(map[string]*Snapshot)}

	for _, name := range names {
		snap, err := loadSnapshot(name)
		if err != nil {
			return nil, err
		}
		a.snapshots[name] = snap
	}

	// Resolve pin.
	if pin == "" || pin == "latest" {
		a.active = names[len(names)-1] // alphabetically last = latest
	} else {
		if _, ok := a.snapshots[pin]; !ok {
			return nil, fmt.Errorf("%w: %q (available: %v)", ErrBadPin, pin, names)
		}
		a.active = pin
	}

	return a, nil
}

// Pin returns the active snapshot name.
func (a *Arsenal) Pin() string { return a.active }

// Available returns all snapshot names.
func (a *Arsenal) Available() []string {
	names := make([]string, 0, len(a.snapshots))
	for name := range a.snapshots {
		names = append(names, name)
	}
	return names
}

// Pick performs an imperative 1:1 map lookup. Returns the model resolved
// through the given source, with source modifiers applied.
func (a *Arsenal) Pick(modelID, sourceID string) (ResolvedAgent, error) {
	snap := a.snapshots[a.active]

	model, ok := snap.Models[modelID]
	if !ok {
		return ResolvedAgent{}, fmt.Errorf("%w: model %q", ErrNotFound, modelID)
	}

	source, ok := snap.Sources[sourceID]
	if !ok {
		return ResolvedAgent{}, fmt.Errorf("%w: source %q", ErrNotFound, sourceID)
	}

	if !canAccess(source, modelID) {
		return ResolvedAgent{}, fmt.Errorf("%w: source %q cannot access model %q", ErrNotFound, sourceID, modelID)
	}

	return resolve(model, source), nil
}

// Select performs declarative intent-based selection. Filters, gates,
// scores, ranks, then exits through Pick(). The intent parameter is
// reserved for future trait inference — currently unused.
func (a *Arsenal) Select(_ string, prefs *Preferences) (ResolvedAgent, error) {
	snap := a.snapshots[a.active]

	type candidate struct {
		model  *ModelEntry
		source *SourceEntry
		score  float64
	}

	var candidates []candidate

	for _, source := range snap.Sources {
		// Source filter.
		if !prefs.Sources.matches(source.Source) {
			continue
		}

		// Iterate models this source can access.
		for modelID, model := range snap.Models {
			if !canAccess(source, modelID) {
				continue
			}

			// Provider filter.
			if !prefs.Providers.matches(model.Provider) {
				continue
			}

			// Model filter.
			if !prefs.Models.matches(model.ID) {
				continue
			}

			// Cost ceiling.
			if prefs.MaxCost > 0 && model.Cost.InputPerM > prefs.MaxCost {
				continue
			}

			// Min traits gate.
			if !model.Traits.meetsMinimum(prefs.MinTraits) {
				continue
			}

			// Score.
			score := model.Traits.Score(prefs.Weights)
			candidates = append(candidates, candidate{model, source, score})
		}
	}

	if len(candidates) == 0 {
		return ResolvedAgent{}, ErrNoCandidate
	}

	// Rank — highest score wins.
	best := candidates[0]
	for _, c := range candidates[1:] {
		if c.score > best.score {
			best = c
		}
	}

	return resolve(best.model, best.source), nil
}

// resolve applies source modifiers to a model and returns a ResolvedAgent.
func resolve(model *ModelEntry, source *SourceEntry) ResolvedAgent {
	effContext := model.Context
	if source.Mods.ContextCap > 0 && source.Mods.ContextCap < effContext {
		effContext = source.Mods.ContextCap
	}

	overhead := source.Mods.TokenOverhead
	if overhead == 0 {
		overhead = 1.0
	}

	pipeline := source.Mods.Pipeline
	if pipeline == "" {
		pipeline = "direct"
	}

	return ResolvedAgent{
		Model:      model.ID,
		Provider:   model.Provider,
		Source:     source.Source,
		Traits:     model.Traits,
		EffContext: effContext,
		Overhead:   overhead,
		Pipeline:   pipeline,
		Cost:       model.Cost,
	}
}

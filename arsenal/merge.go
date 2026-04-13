// merge.go — merge live discovery results with static YAML catalog.
//
// Live API → authoritative for availability + context window.
// Static YAML → authoritative for benchmarks + traits + cost.
// Unknown models from API → added with zero traits.
// Models in YAML but not in API → marked unavailable.
//
// TRP-TSK-130, TRP-GOL-8
package arsenal

// MergeDiscovery updates the snapshot's model entries with live discovery data.
// Models matched by ID get their Available flag and ContextWindow updated.
// Unknown models from the API are added with zero traits.
// Models in the catalog but NOT in the discovery results are marked unavailable.
func MergeDiscovery(snap *Snapshot, discovered []DiscoveredModel) {
	if len(discovered) == 0 {
		return
	}
	snap.discoveryRan = true

	// Build set of discovered model IDs per provider.
	seen := make(map[string]DiscoveredModel, len(discovered))
	for _, d := range discovered {
		seen[d.ID] = d
	}

	// Update existing models.
	for id, model := range snap.Models {
		if d, ok := seen[id]; ok {
			model.Available = d.Available
			if d.ContextWindow > 0 {
				model.Context = d.ContextWindow
			}
			snap.Models[id] = model
			delete(seen, id) // consumed
		} else if model.Provider == discovered[0].Provider {
			// Same provider but not in API → unavailable.
			model.Available = false
			snap.Models[id] = model
		}
	}

	// Add unknown models from API (new models not yet in YAML catalog).
	for _, d := range seen {
		if !d.Available {
			continue // skip unavailable unknown models
		}
		snap.Models[d.ID] = &ModelEntry{
			ID:        d.ID,
			Provider:  d.Provider,
			Context:   d.ContextWindow,
			Available: true,
			// Zero traits — no benchmarks yet for this new model.
		}
	}
}

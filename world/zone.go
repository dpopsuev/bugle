package world

import "sync"

// ZonedWorld partitions entities into named zones for locality-aware queries.
// The underlying World is unchanged — zones are a routing layer on top.
type ZonedWorld struct {
	world *World
	mu    sync.RWMutex
	zones map[string]map[EntityID]bool // zone name → entity IDs
}

// NewZonedWorld wraps a World with zone-based partitioning.
func NewZonedWorld(w *World) *ZonedWorld {
	return &ZonedWorld{
		world: w,
		zones: make(map[string]map[EntityID]bool),
	}
}

// World returns the underlying World.
func (z *ZonedWorld) World() *World { return z.world }

// SpawnInZone creates a new entity and assigns it to the given zone.
func (z *ZonedWorld) SpawnInZone(zone string) EntityID {
	id := z.world.Spawn()
	z.mu.Lock()
	if z.zones[zone] == nil {
		z.zones[zone] = make(map[EntityID]bool)
	}
	z.zones[zone][id] = true
	z.mu.Unlock()
	return id
}

// ZoneOf returns the zone an entity belongs to, or "" if unzoned.
func (z *ZonedWorld) ZoneOf(id EntityID) string {
	z.mu.RLock()
	defer z.mu.RUnlock()
	for zone, entities := range z.zones {
		if entities[id] {
			return zone
		}
	}
	return ""
}

// QueryZone returns all entities in a zone that have the given component type.
func (z *ZonedWorld) QueryZone(zone string, ct ComponentType) []EntityID {
	z.mu.RLock()
	zoneEntities := z.zones[zone]
	z.mu.RUnlock()

	if len(zoneEntities) == 0 {
		return nil
	}

	all := z.world.QueryType(ct)
	result := make([]EntityID, 0, len(all))
	for _, id := range all {
		if zoneEntities[id] {
			result = append(result, id)
		}
	}
	return result
}

// Zones returns all zone names.
func (z *ZonedWorld) Zones() []string {
	z.mu.RLock()
	defer z.mu.RUnlock()
	names := make([]string, 0, len(z.zones))
	for name := range z.zones {
		names = append(names, name)
	}
	return names
}

// Remove takes an entity out of its zone (does not despawn).
func (z *ZonedWorld) Remove(id EntityID) {
	z.mu.Lock()
	defer z.mu.Unlock()
	for _, entities := range z.zones {
		delete(entities, id)
	}
}

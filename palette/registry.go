package palette

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"sync"
)

// Sentinel errors for palette operations.
var (
	ErrAllSlotsAssigned = errors.New("palette: all 56 color slots are assigned")
	ErrUnknownShade     = errors.New("palette: unknown shade")
	ErrShadeExhausted   = errors.New("palette: all colors in shade are assigned")
	ErrUnknownColor     = errors.New("palette: unknown color")
	ErrColorWrongShade  = errors.New("palette: color belongs to different shade")
	ErrAlreadyAssigned  = errors.New("palette: color already assigned")
)

// Registry manages color identity assignment with collision prevention.
type Registry struct {
	mu       sync.Mutex
	assigned map[string]bool // "shade·color" → true
}

// NewRegistry creates an empty identity registry.
func NewRegistry() *Registry {
	return &Registry{
		assigned: make(map[string]bool),
	}
}

func registryKey(shade, color string) string {
	return shade + "·" + color
}

// Assign returns a ColorIdentity with a random available color.
// Picks a random shade, then a random available color within it.
func (r *Registry) Assign(role, collective string) (ColorIdentity, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Shuffle shade order for randomness
	shades := make([]int, len(Palette))
	for i := range shades {
		shades[i] = i
	}
	rand.Shuffle(len(shades), func(i, j int) { shades[i], shades[j] = shades[j], shades[i] })

	for _, si := range shades {
		shade := Palette[si]
		for _, color := range shade.Colors {
			key := registryKey(shade.Name, color.Name)
			if !r.assigned[key] {
				r.assigned[key] = true
				return ColorIdentity{
					Shade:      shade.Name,
					Color:      color.Name,
					Role:       role,
					Collective: collective,
					Hex:        color.Hex,
				}, nil
			}
		}
	}
	return ColorIdentity{}, ErrAllSlotsAssigned
}

// AssignInGroup returns a ColorIdentity from a specific shade family.
func (r *Registry) AssignInGroup(shade, role, collective string) (ColorIdentity, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	s := LookupShade(shade)
	if s == nil {
		return ColorIdentity{}, fmt.Errorf("%w: %q", ErrUnknownShade, shade)
	}

	for _, color := range s.Colors {
		key := registryKey(s.Name, color.Name)
		if !r.assigned[key] {
			r.assigned[key] = true
			return ColorIdentity{
				Shade:      s.Name,
				Color:      color.Name,
				Role:       role,
				Collective: collective,
				Hex:        color.Hex,
			}, nil
		}
	}
	return ColorIdentity{}, fmt.Errorf("%w: %q", ErrShadeExhausted, shade)
}

// Set explicitly assigns a specific shade+color combination.
func (r *Registry) Set(shade, color, role, collective string) (ColorIdentity, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	c, foundShade, ok := LookupColor(color)
	if !ok {
		return ColorIdentity{}, fmt.Errorf("%w: %q", ErrUnknownColor, color)
	}
	if foundShade != shade {
		return ColorIdentity{}, fmt.Errorf("%w: %q belongs to shade %q, not %q", ErrColorWrongShade, color, foundShade, shade)
	}

	key := registryKey(shade, color)
	if r.assigned[key] {
		return ColorIdentity{}, fmt.Errorf("%w: %s·%s", ErrAlreadyAssigned, shade, color)
	}

	r.assigned[key] = true
	return ColorIdentity{
		Shade:      shade,
		Color:      c.Name,
		Role:       role,
		Collective: collective,
		Hex:        c.Hex,
	}, nil
}

// Release returns a color to the available pool.
func (r *Registry) Release(id ColorIdentity) { //nolint:gocritic // value param for API simplicity
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.assigned, registryKey(id.Shade, id.Color))
}

// Active returns all currently assigned identities' keys.
func (r *Registry) Active() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.assigned)
}

// Package symbol provides visual identity primitives for agents.
// Color, Element, Persona — the cosmetic layer, always rendered.
// Absorbs palette/, element/, identity/ from pre-v0.2.0.
package symbol

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"sync"

	"github.com/dpopsuev/jericho/world"
)

// ColorType is the ComponentType for Color (visual identity ECS component).
const ColorType world.ComponentType = "color"

// Shade is a color family grouping for agent collectives.
type Shade struct {
	Name   string
	Colors []PaletteColor
}

// PaletteColor is a specific color within a shade family.
type PaletteColor struct {
	Name string
	Hex  string
}

// Color is the visual identity ECS component for agents.
// Format: "Denim Writer of Indigo Refactor" (Color Role of Shade Collective).
type Color struct {
	Shade      string `json:"shade"`      // group family: "Indigo", "Crimson"
	Name       string `json:"name"`       // individual: "Denim", "Scarlet"
	Role       string `json:"role"`       // function: "Writer", "Reviewer"
	Collective string `json:"collective"` // formation: "Refactor", "Triage"
	Hex        string `json:"hex"`        // CSS hex: "#6F8FAF"
}

// ComponentType implements world.Component.
func (Color) ComponentType() world.ComponentType { return ColorType }

// Title returns the heraldic name: "Denim Writer of Indigo Refactor".
func (c Color) Title() string { //nolint:gocritic // value receiver needed for ECS Get[T]
	return fmt.Sprintf("%s %s of %s %s", c.Name, c.Role, c.Shade, c.Collective)
}

// Label returns the compact log format: "[Indigo·Denim|Writer]".
func (c Color) Label() string { //nolint:gocritic // value receiver needed for ECS Get[T]
	return fmt.Sprintf("[%s·%s|%s]", c.Shade, c.Name, c.Role)
}

// Short returns just the color name: "Denim".
func (c Color) Short() string { return c.Name } //nolint:gocritic // value receiver

// ContrastMode indicates whether the terminal uses light or dark background.
type ContrastMode string

const (
	ContrastAuto  ContrastMode = "auto" // detect from terminal
	ContrastDark  ContrastMode = "dark"
	ContrastLight ContrastMode = "light"
)

// ANSI returns a 24-bit true color ANSI escape sequence for foreground text.
func (c Color) ANSI() string { //nolint:gocritic // value receiver for ECS
	if len(c.Hex) != 7 {
		return ""
	}
	var r, g, b uint8
	fmt.Sscanf(c.Hex, "#%02x%02x%02x", &r, &g, &b) //nolint:errcheck // hex format guaranteed by palette
	return fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b)
}

// Reservation is a color preference, not an assignment.
// Used by persona templates to express preferred shade/color without
// locking in a specific assignment (the Registry handles collisions).
type Reservation struct {
	Shade string // preferred shade family (empty = any)
	Color string // preferred color (empty = any in shade)
}

// Palette defines 7 shade families x 8 colors = 56 unique agent identities.
var Palette = []Shade{
	{Name: "Azure", Colors: []PaletteColor{
		{"Cerulean", "#007BA7"},
		{"Cobalt", "#0047AB"},
		{"Sapphire", "#0F52BA"},
		{"Indigo", "#4B0082"},
		{"Navy", "#000080"},
		{"Periwinkle", "#CCCCFF"},
		{"Steel", "#4682B4"},
		{"Teal", "#008080"},
	}},
	{Name: "Crimson", Colors: []PaletteColor{
		{"Scarlet", "#FF2400"},
		{"Vermillion", "#E34234"},
		{"Ruby", "#E0115F"},
		{"Garnet", "#733635"},
		{"Cardinal", "#C41E3A"},
		{"Carmine", "#960018"},
		{"Rust", "#B7410E"},
		{"Coral", "#FF7F50"},
	}},
	{Name: "Forest", Colors: []PaletteColor{
		{"Emerald", "#50C878"},
		{"Jade", "#00A86B"},
		{"Sage", "#BCB88A"},
		{"Olive", "#808000"},
		{"Mint", "#3EB489"},
		{"Hunter", "#355E3B"},
		{"Moss", "#8A9A5B"},
		{"Viridian", "#40826D"},
	}},
	{Name: "Amber", Colors: []PaletteColor{
		{"Saffron", "#F4C430"},
		{"Gold", "#FFD700"},
		{"Marigold", "#EAA221"},
		{"Tangerine", "#FF9966"},
		{"Apricot", "#FBCEB1"},
		{"Ochre", "#CC7722"},
		{"Bronze", "#CD7F32"},
		{"Copper", "#B87333"},
	}},
	{Name: "Violet", Colors: []PaletteColor{
		{"Amethyst", "#9966CC"},
		{"Lavender", "#E6E6FA"},
		{"Plum", "#8E4585"},
		{"Mauve", "#E0B0FF"},
		{"Orchid", "#DA70D6"},
		{"Thistle", "#D8BFD8"},
		{"Iris", "#5A4FCF"},
		{"Heather", "#B7C3D0"},
	}},
	{Name: "Slate", Colors: []PaletteColor{
		{"Charcoal", "#36454F"},
		{"Ash", "#B2BEB5"},
		{"Pewter", "#8BA8B7"},
		{"Silver", "#C0C0C0"},
		{"Smoke", "#738276"},
		{"Graphite", "#383838"},
		{"Iron", "#48494B"},
		{"Flint", "#6F6A63"},
	}},
	{Name: "Ivory", Colors: []PaletteColor{
		{"Pearl", "#EAE0C8"},
		{"Cream", "#FFFDD0"},
		{"Linen", "#FAF0E6"},
		{"Snow", "#FFFAFA"},
		{"Alabaster", "#F2F0E6"},
		{"Bone", "#E3DAC9"},
		{"Shell", "#FFF5EE"},
		{"Chalk", "#FDFDFD"},
	}},
	// Reserved shades — 5 empty slots for consumer customization.
	{Name: "Teal", Colors: []PaletteColor{
		{"Aquamarine", "#7FFFD4"},
		{"Turquoise", "#40E0D0"},
		{"Cyan", "#00FFFF"},
		{"Arctic", "#4CC8DB"},
		{"Lagoon", "#4CB7A5"},
		{"Seafoam", "#71D9B7"},
		{"Reef", "#009B8D"},
		{"Marina", "#4CBFA6"},
	}},
	{Name: "Rose", Colors: []PaletteColor{
		{"Blush", "#DE5D83"},
		{"Peony", "#DB7093"},
		{"Fuchsia", "#FF00FF"},
		{"Magenta", "#FF0090"},
		{"Cerise", "#DE3163"},
		{"Petal", "#F4C2C2"},
		{"Blossom", "#FFB7C5"},
		{"Rosewood", "#65000B"},
	}},
	{Name: "Bronze", Colors: []PaletteColor{
		{"Umber", "#635147"},
		{"Sienna", "#A0522D"},
		{"Mahogany", "#C04000"},
		{"Chestnut", "#954535"},
		{"Walnut", "#773F1A"},
		{"Cinnamon", "#D2691E"},
		{"Toffee", "#755139"},
		{"Espresso", "#3C1414"},
	}},
	{Name: "Indigo", Colors: []PaletteColor{
		{"Midnight", "#191970"},
		{"Dusk", "#7B68EE"},
		{"Twilight", "#5B5EA6"},
		{"Nebula", "#483D8B"},
		{"Eclipse", "#3C1361"},
		{"Cosmos", "#443399"},
		{"Galactic", "#2E2D88"},
		{"Aurora", "#6A5ACD"},
	}},
	{Name: "Sage", Colors: []PaletteColor{
		{"Laurel", "#A9BA9D"},
		{"Pistachio", "#93C572"},
		{"Lichen", "#7DA98D"},
		{"Fern", "#4F7942"},
		{"Thyme", "#596D48"},
		{"Basil", "#5C7A29"},
		{"Clover", "#3EA055"},
		{"Willow", "#65A374"},
	}},
}

// LookupShade finds a shade by name. Returns nil if not found.
func LookupShade(name string) *Shade {
	for i := range Palette {
		if Palette[i].Name == name {
			return &Palette[i]
		}
	}
	return nil
}

// LookupColor finds a color by name across all shades.
// Returns the color and its parent shade name.
func LookupColor(name string) (PaletteColor, string, bool) {
	for _, shade := range Palette {
		for _, c := range shade.Colors {
			if c.Name == name {
				return c, shade.Name, true
			}
		}
	}
	return PaletteColor{}, "", false
}

// Sentinel errors for color registry operations.
var (
	ErrAllSlotsAssigned = errors.New("symbol: all 56 color slots are assigned")
	ErrUnknownShade     = errors.New("symbol: unknown shade")
	ErrShadeExhausted   = errors.New("symbol: all colors in shade are assigned")
	ErrUnknownColor     = errors.New("symbol: unknown color")
	ErrColorWrongShade  = errors.New("symbol: color belongs to different shade")
	ErrAlreadyAssigned  = errors.New("symbol: color already assigned")
)

// Registry manages color identity assignment with collision prevention.
type Registry struct {
	mu       sync.Mutex
	assigned map[string]bool // "shade·color" → true
}

// NewRegistry creates an empty color registry.
func NewRegistry() *Registry {
	return &Registry{
		assigned: make(map[string]bool),
	}
}

func registryKey(shade, color string) string {
	return shade + "·" + color
}

// Assign returns a Color with a random available color.
func (r *Registry) Assign(role, collective string) (Color, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

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
				return Color{
					Shade:      shade.Name,
					Name:       color.Name,
					Role:       role,
					Collective: collective,
					Hex:        color.Hex,
				}, nil
			}
		}
	}
	return Color{}, ErrAllSlotsAssigned
}

// AssignInGroup returns a Color from a specific shade family.
func (r *Registry) AssignInGroup(shade, role, collective string) (Color, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	s := LookupShade(shade)
	if s == nil {
		return Color{}, fmt.Errorf("%w: %q", ErrUnknownShade, shade)
	}

	for _, color := range s.Colors {
		key := registryKey(s.Name, color.Name)
		if !r.assigned[key] {
			r.assigned[key] = true
			return Color{
				Shade:      s.Name,
				Name:       color.Name,
				Role:       role,
				Collective: collective,
				Hex:        color.Hex,
			}, nil
		}
	}
	return Color{}, fmt.Errorf("%w: %q", ErrShadeExhausted, shade)
}

// AssignWithPreference tries the preferred shade+color, falls back if taken.
func (r *Registry) AssignWithPreference(res Reservation, role, collective string) (Color, error) {
	if res.Shade != "" && res.Color != "" {
		c, err := r.Set(res.Shade, res.Color, role, collective)
		if err == nil {
			return c, nil
		}
	}
	if res.Shade != "" {
		return r.AssignInGroup(res.Shade, role, collective)
	}
	return r.Assign(role, collective)
}

// Set explicitly assigns a specific shade+color combination.
func (r *Registry) Set(shade, color, role, collective string) (Color, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	c, foundShade, ok := LookupColor(color)
	if !ok {
		return Color{}, fmt.Errorf("%w: %q", ErrUnknownColor, color)
	}
	if foundShade != shade {
		return Color{}, fmt.Errorf("%w: %q belongs to shade %q, not %q", ErrColorWrongShade, color, foundShade, shade)
	}

	key := registryKey(shade, color)
	if r.assigned[key] {
		return Color{}, fmt.Errorf("%w: %s·%s", ErrAlreadyAssigned, shade, color)
	}

	r.assigned[key] = true
	return Color{
		Shade:      shade,
		Name:       c.Name,
		Role:       role,
		Collective: collective,
		Hex:        c.Hex,
	}, nil
}

// Release returns a color to the available pool.
func (r *Registry) Release(c Color) { //nolint:gocritic // value param for API simplicity
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.assigned, registryKey(c.Shade, c.Name))
}

// Active returns the count of currently assigned colors.
func (r *Registry) Active() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.assigned)
}

package identity

import (
	"fmt"
	"strings"
)

// Archetype is a perennial agent identity template — stable across model
// releases while the models behind them shift. Flat struct replaces the
// old AgentIdentity monolith with direct fields.
type Archetype struct {
	Name            string             `json:"name"`
	Description     string             `json:"description"`
	ColorPref       Reservation        `json:"color_pref"`       // preferred color (not assigned)
	Element         Element            `json:"element"`          // behavioral archetype
	Position        Position           `json:"position"`         // dialectic structural role
	Alignment       Alignment          `json:"alignment"`        // thesis or antithesis
	HomeZone        MetaPhase          `json:"home_zone"`        // circuit zone
	StickinessLevel int                `json:"stickiness_level"` // zone affinity
	Role            Role               `json:"role,omitempty"`   // org role
	StepAffinity    map[string]float64 `json:"step_affinity,omitempty"`
	PersonalityTags []string           `json:"personality_tags,omitempty"`
	PromptPreamble  string             `json:"prompt_preamble,omitempty"`
}

// Tag returns a log-friendly tag like "[iron/judge]".
func (a Archetype) Tag() string { //nolint:gocritic // value receiver for log formatting
	color := strings.ToLower(a.ColorPref.Color)
	if color == "" {
		color = "none"
	}
	name := strings.ToLower(a.Name)
	if name == "" {
		name = "anon"
	}
	return fmt.Sprintf("[%s/%s]", color, name)
}

// ArchetypeResolver looks up a persona by name.
type ArchetypeResolver func(name string) (Archetype, bool) //nolint:revive // kept for Origami alias compat

// DefaultArchetypeResolver is the active persona lookup function. Nil until
// a persona package registers itself via init().
var DefaultArchetypeResolver ArchetypeResolver //nolint:revive // kept for Origami alias compat

// Alignment represents an agent's motivational orientation.
type Alignment string

const (
	AlignmentThesis     Alignment = "thesis"
	AlignmentAntithesis Alignment = "antithesis"
)

// Position represents an agent's dialectic position (structural role).
type Position string

const (
	PositionPG Position = "PG"
	PositionSG Position = "SG"
	PositionPF Position = "PF"
	PositionC  Position = "C"
)

// MetaPhase represents a zone in the circuit graph.
type MetaPhase string

const (
	MetaPhaseBk MetaPhase = "Backcourt"
	MetaPhaseFc MetaPhase = "Frontcourt"
	MetaPhasePt MetaPhase = "Paint"
)

// Role represents an agent's organizational role in an agentic hierarchy.
type Role string

const (
	RoleWorker   Role = "worker"
	RoleManager  Role = "manager"
	RoleEnforcer Role = "enforcer"
	RoleBroker   Role = "broker"
)

// ValidRoles contains all recognized role values for validation.
var ValidRoles = map[Role]bool{
	RoleWorker:   true,
	RoleManager:  true,
	RoleEnforcer: true,
	RoleBroker:   true,
}

// HomeZoneFor returns the MetaPhase for a given Position.
func HomeZoneFor(p Position) MetaPhase {
	switch p {
	case PositionPG:
		return MetaPhaseBk
	case PositionSG:
		return MetaPhasePt
	case PositionPF:
		return MetaPhaseFc
	case PositionC:
		return MetaPhaseFc
	default:
		return ""
	}
}

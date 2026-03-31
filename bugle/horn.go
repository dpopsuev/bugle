package bugle

// HornLevel represents health severity — the bugle calls.
type HornLevel string

const (
	HornGreen  HornLevel = "green"  // Silence — all nominal.
	HornYellow HornLevel = "yellow" // Warning call — approaching limits.
	HornRed    HornLevel = "red"    // Alarm call — limits exceeded, stalled.
	HornBlack  HornLevel = "black"  // The walls have fallen — unrecoverable.
)

// HornCategory classifies what triggered the horn.
type HornCategory string

const (
	CategoryBudget    HornCategory = "budget"
	CategoryDeadlock  HornCategory = "deadlock"
	CategoryLifecycle HornCategory = "lifecycle"
	CategoryQuality   HornCategory = "quality"
	CategorySecurity  HornCategory = "security"
	CategoryDrift     HornCategory = "drift"
)

// Horn is a health signal reported by a worker or aggregated by a server.
type Horn struct {
	Level    HornLevel    `json:"level"`
	Category HornCategory `json:"category,omitempty"`
	Message  string       `json:"message,omitempty"`
}

// Worse returns true if h is a more severe horn than other.
func (h HornLevel) Worse(other HornLevel) bool {
	return hornOrdinal(h) > hornOrdinal(other)
}

// WorstHorn returns the most severe horn level from a set.
func WorstHorn(levels ...HornLevel) HornLevel {
	worst := HornGreen
	for _, l := range levels {
		if l.Worse(worst) {
			worst = l
		}
	}
	return worst
}

func hornOrdinal(l HornLevel) int {
	switch l {
	case HornGreen:
		return 0
	case HornYellow:
		return 1
	case HornRed:
		return 2
	case HornBlack:
		return 3
	default:
		return -1
	}
}

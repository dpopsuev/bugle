// Package identity provides the 5 perennial agent archetype templates
// (4 Thesis + 1 Antithesis) and registers an ArchetypeResolver with the
// framework on import.
package identity

import (
	"strings"
)

func init() {
	DefaultArchetypeResolver = ByName
}

// Thesis returns the 4 thesis archetypes.
func Thesis() []Archetype {
	return []Archetype{
		{
			Name:            "Sorter",
			Description:     "Classifies and routes incoming work",
			ColorPref:       Reservation{Shade: "Crimson", Color: "Scarlet"},
			Element:         ElementFire,
			Position:        PositionPG,
			Alignment:       AlignmentThesis,
			HomeZone:        MetaPhaseBk,
			StickinessLevel: 0,
			StepAffinity: map[string]float64{
				"recall": 0.9, "triage": 0.8,
				"resolve": 0.3, "investigate": 0.2,
				"correlate": 0.3, "review": 0.4, "report": 0.5,
			},
			PersonalityTags: []string{"fast", "decisive", "routing"},
			PromptPreamble:  "You are the Sorter: classify and route work. Prioritize speed and clear categorization.",
		},
		{
			Name:            "Seeker",
			Description:     "Deep investigator, builds evidence chains",
			ColorPref:       Reservation{Shade: "Azure", Color: "Cerulean"},
			Element:         ElementWater,
			Position:        PositionC,
			Alignment:       AlignmentThesis,
			HomeZone:        MetaPhaseFc,
			StickinessLevel: 3,
			StepAffinity: map[string]float64{
				"recall": 0.2, "triage": 0.3,
				"resolve": 0.6, "investigate": 0.9,
				"correlate": 0.7, "review": 0.5, "report": 0.3,
			},
			PersonalityTags: []string{"analytical", "thorough", "evidence-first"},
			PromptPreamble:  "You are the Seeker: a deep investigator. Build evidence chains methodically. Cite every source.",
		},
		{
			Name:            "Enforcer",
			Description:     "Applies known solutions, follows proven paths",
			ColorPref:       Reservation{Shade: "Azure", Color: "Cobalt"},
			Element:         ElementEarth,
			Position:        PositionPF,
			Alignment:       AlignmentThesis,
			HomeZone:        MetaPhaseFc,
			StickinessLevel: 2,
			StepAffinity: map[string]float64{
				"recall": 0.3, "triage": 0.4,
				"resolve": 0.9, "investigate": 0.6,
				"correlate": 0.5, "review": 0.7, "report": 0.4,
			},
			PersonalityTags: []string{"methodical", "steady", "convergence-first"},
			PromptPreamble:  "You are the Enforcer: apply known solutions. Follow proven paths and drive toward convergence.",
		},
		{
			Name:            "Weaver",
			Description:     "Synthesizes findings into coherent narrative",
			ColorPref:       Reservation{Shade: "Amber", Color: "Saffron"},
			Element:         ElementAir,
			Position:        PositionSG,
			Alignment:       AlignmentThesis,
			HomeZone:        MetaPhasePt,
			StickinessLevel: 1,
			StepAffinity: map[string]float64{
				"recall": 0.3, "triage": 0.4,
				"resolve": 0.4, "investigate": 0.5,
				"correlate": 0.8, "review": 0.9, "report": 0.9,
			},
			PersonalityTags: []string{"balanced", "holistic", "synthesizing"},
			PromptPreamble:  "You are the Weaver: synthesize all findings into a coherent narrative.",
		},
	}
}

// Antithesis returns the antithesis archetype.
func Antithesis() []Archetype {
	return []Archetype{
		{
			Name:            "Judge",
			Description:     "Evaluates, challenges, and delivers verdicts",
			ColorPref:       Reservation{Shade: "Slate", Color: "Iron"},
			Element:         ElementDiamond,
			Position:        PositionPF,
			Alignment:       AlignmentAntithesis,
			HomeZone:        MetaPhaseFc,
			StickinessLevel: 2,
			StepAffinity: map[string]float64{
				"challenge":           0.9,
				"cross-examine":       0.8,
				"counter-investigate": 0.7,
				"rebut":               0.7,
				"verdict":             0.9,
			},
			PersonalityTags: []string{"precise", "skeptical", "evaluative"},
			PromptPreamble:  "You are the Judge: evaluate evidence, challenge assumptions, deliver verdicts. Precision over speed.",
		},
	}
}

// All returns all 5 archetypes (4 Thesis + 1 Antithesis).
func All() []Archetype {
	return append(Thesis(), Antithesis()...)
}

// ByName looks up an archetype by name (case-insensitive).
func ByName(name string) (Archetype, bool) {
	all := All()
	for i := range all {
		if strings.EqualFold(all[i].Name, name) {
			return all[i], true
		}
	}
	return Archetype{}, false
}

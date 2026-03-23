package element

import "testing"

func TestAllElements_Count(t *testing.T) {
	if got := len(AllElements()); got != 6 {
		t.Errorf("AllElements() = %d, want 6", got)
	}
}

func TestAllApproaches_Count(t *testing.T) {
	if got := len(AllApproaches()); got != 6 {
		t.Errorf("AllApproaches() = %d, want 6", got)
	}
}

func TestDefaultTraits_AllElements(t *testing.T) {
	for _, e := range AllElements() {
		traits := DefaultTraits(e)
		if traits.Element != e {
			t.Errorf("DefaultTraits(%s).Element = %s", e, traits.Element)
		}
		if traits.Speed == "" {
			t.Errorf("DefaultTraits(%s).Speed is empty", e)
		}
		if traits.MaxLoops < 0 {
			t.Errorf("DefaultTraits(%s).MaxLoops = %d, want >= 0", e, traits.MaxLoops)
		}
	}
}

func TestResolveApproach_AllApproaches(t *testing.T) {
	for _, a := range AllApproaches() {
		e, ok := ResolveApproach(string(a))
		if !ok {
			t.Errorf("ResolveApproach(%s) = false", a)
		}
		if e == "" {
			t.Errorf("ResolveApproach(%s) returned empty element", a)
		}
	}
}

func TestResolveApproach_Unknown(t *testing.T) {
	_, ok := ResolveApproach("nonexistent")
	if ok {
		t.Error("ResolveApproach(nonexistent) should return false")
	}
}

func TestApproachForElement_RoundTrip(t *testing.T) {
	for _, e := range AllElements() {
		a := ApproachForElement(e)
		back, ok := ResolveApproach(string(a))
		if !ok {
			t.Errorf("ApproachForElement(%s) = %s, ResolveApproach failed", e, a)
		}
		if back != e {
			t.Errorf("round-trip: %s → %s → %s", e, a, back)
		}
	}
}

func TestApproachEmoji_NonEmpty(t *testing.T) {
	for _, a := range AllApproaches() {
		if emoji := ApproachEmoji(a); emoji == "" {
			t.Errorf("ApproachEmoji(%s) is empty", a)
		}
	}
}

func TestApproachTraits_AllApproaches(t *testing.T) {
	for _, a := range AllApproaches() {
		traits := ApproachTraits(a)
		if traits.Element == "" {
			t.Errorf("ApproachTraits(%s).Element is empty", a)
		}
	}
}

func TestApproachTraitsSummary_NonEmpty(t *testing.T) {
	for _, a := range AllApproaches() {
		s := ApproachTraitsSummary(a)
		if s == "" {
			t.Errorf("ApproachTraitsSummary(%s) is empty", a)
		}
	}
}

func TestElementConstants(t *testing.T) {
	elements := []Element{ElementFire, ElementLightning, ElementEarth, ElementDiamond, ElementWater, ElementAir}
	if len(elements) != 6 {
		t.Fatalf("expected 6 element constants, got %d", len(elements))
	}
	seen := make(map[Element]bool)
	for _, e := range elements {
		if seen[e] {
			t.Errorf("duplicate element constant: %s", e)
		}
		seen[e] = true
	}
}

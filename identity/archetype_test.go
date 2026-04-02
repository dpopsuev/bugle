package identity

import (
	"testing"
)

func TestAll_Count(t *testing.T) {
	all := All()
	if len(all) != 8 {
		t.Errorf("len(All) = %d, want 8", len(all))
	}
}

func TestThesis_Count(t *testing.T) {
	thesis := Thesis()
	if len(thesis) != 4 {
		t.Errorf("len(Thesis) = %d, want 4", len(thesis))
	}
	for _, p := range thesis {
		if p.Alignment != AlignmentThesis {
			t.Errorf("persona %q has alignment %q, want thesis", p.Name, p.Alignment)
		}
	}
}

func TestAntithesis_Count(t *testing.T) {
	antithesis := Antithesis()
	if len(antithesis) != 4 {
		t.Errorf("len(Antithesis) = %d, want 4", len(antithesis))
	}
	for _, p := range antithesis {
		if p.Alignment != AlignmentAntithesis {
			t.Errorf("persona %q has alignment %q, want antithesis", p.Name, p.Alignment)
		}
	}
}

func TestByName_Herald(t *testing.T) {
	p, ok := ByName("Herald")
	if !ok {
		t.Fatal("ByName(Herald) not found")
	}
	if p.ColorPref.Color != "Scarlet" {
		t.Errorf("Herald color pref = %q, want Scarlet", p.ColorPref.Color)
	}
	if p.Element != ElementFire {
		t.Errorf("Herald element = %q, want fire", p.Element)
	}
	if p.Position != PositionPG {
		t.Errorf("Herald position = %q, want PG", p.Position)
	}
	if p.Alignment != AlignmentThesis {
		t.Errorf("Herald alignment = %q, want thesis", p.Alignment)
	}
}

func TestByName_CaseInsensitive(t *testing.T) {
	_, ok := ByName("herald")
	if !ok {
		t.Error("ByName should be case-insensitive")
	}
	_, ok = ByName("CHALLENGER")
	if !ok {
		t.Error("ByName should be case-insensitive")
	}
}

func TestByName_NotFound(t *testing.T) {
	_, ok := ByName("nonexistent")
	if ok {
		t.Error("ByName should return false for nonexistent name")
	}
}

func TestPersonas_UniqueNames(t *testing.T) {
	all := All()
	seen := make(map[string]bool, len(all))
	for _, p := range all {
		if seen[p.Name] {
			t.Errorf("duplicate persona name: %s", p.Name)
		}
		seen[p.Name] = true
	}
}

func TestPersonas_AllPositionsCovered(t *testing.T) {
	positions := map[Position]int{PositionPG: 0, PositionSG: 0, PositionPF: 0, PositionC: 0}
	for _, p := range All() {
		positions[p.Position]++
	}
	for pos, count := range positions {
		if count != 2 {
			t.Errorf("position %s has %d personas, want 2 (1 thesis + 1 antithesis)", pos, count)
		}
	}
}

func TestPersonas_AllHaveStepAffinity(t *testing.T) {
	for _, p := range All() {
		if len(p.StepAffinity) == 0 {
			t.Errorf("persona %s has no step affinity", p.Name)
		}
	}
}

func TestPersonas_AllHavePromptPreamble(t *testing.T) {
	for _, p := range All() {
		if p.PromptPreamble == "" {
			t.Errorf("persona %s has empty prompt preamble", p.Name)
		}
	}
}

func TestPersonas_HomeZoneMatchesPosition(t *testing.T) {
	for _, p := range All() {
		expected := HomeZoneFor(p.Position)
		if p.HomeZone != expected {
			t.Errorf("persona %s: HomeZone=%q but HomeZoneFor(%s)=%q",
				p.Name, p.HomeZone, p.Position, expected)
		}
	}
}

func TestPersonas_AllHaveColorPref(t *testing.T) {
	for _, p := range All() {
		if p.ColorPref.Shade == "" {
			t.Errorf("persona %s has empty color preference shade", p.Name)
		}
		if p.ColorPref.Color == "" {
			t.Errorf("persona %s has empty color preference color", p.Name)
		}
	}
}

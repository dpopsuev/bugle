package identity

import "testing"

func TestAll_Count(t *testing.T) {
	if got := len(All()); got != 5 {
		t.Errorf("All() = %d, want 5", got)
	}
}

func TestThesis_Count(t *testing.T) {
	th := Thesis()
	if len(th) != 4 {
		t.Fatalf("Thesis() = %d, want 4", len(th))
	}
	for _, p := range th {
		if p.Alignment != AlignmentThesis {
			t.Errorf("%s alignment = %s, want thesis", p.Name, p.Alignment)
		}
	}
}

func TestAntithesis_Count(t *testing.T) {
	anti := Antithesis()
	if len(anti) != 1 {
		t.Fatalf("Antithesis() = %d, want 1", len(anti))
	}
	if anti[0].Name != "Judge" {
		t.Errorf("Antithesis name = %s, want Judge", anti[0].Name)
	}
	if anti[0].Alignment != AlignmentAntithesis {
		t.Errorf("Judge alignment = %s, want antithesis", anti[0].Alignment)
	}
}

func TestByName_Sorter(t *testing.T) {
	p, ok := ByName("Sorter")
	if !ok {
		t.Fatal("Sorter not found")
	}
	if p.Element != ElementFire {
		t.Errorf("Sorter element = %s, want Fire", p.Element)
	}
	if p.Position != PositionPG {
		t.Errorf("Sorter position = %s, want PG", p.Position)
	}
}

func TestByName_Judge(t *testing.T) {
	p, ok := ByName("Judge")
	if !ok {
		t.Fatal("Judge not found")
	}
	if p.Element != ElementDiamond {
		t.Errorf("Judge element = %s, want Diamond", p.Element)
	}
	if p.Alignment != AlignmentAntithesis {
		t.Errorf("Judge alignment = %s, want antithesis", p.Alignment)
	}
}

func TestByName_CaseInsensitive(t *testing.T) {
	if _, ok := ByName("SEEKER"); !ok {
		t.Error("case-insensitive lookup for SEEKER failed")
	}
	if _, ok := ByName("judge"); !ok {
		t.Error("case-insensitive lookup for judge failed")
	}
}

func TestByName_OldNamesRemoved(t *testing.T) {
	removed := []string{"Herald", "Sentinel", "Challenger", "Abyss", "Bulwark", "Specter"}
	for _, name := range removed {
		if _, ok := ByName(name); ok {
			t.Errorf("%s should be removed, but was found", name)
		}
	}
}

func TestArchetypes_UniqueNames(t *testing.T) {
	seen := make(map[string]bool)
	for _, p := range All() {
		if seen[p.Name] {
			t.Errorf("duplicate archetype name: %s", p.Name)
		}
		seen[p.Name] = true
	}
}

func TestArchetypes_AllHaveStepAffinity(t *testing.T) {
	for _, p := range All() {
		if len(p.StepAffinity) == 0 {
			t.Errorf("%s has no step affinity", p.Name)
		}
	}
}

func TestArchetypes_AllHavePromptPreamble(t *testing.T) {
	for _, p := range All() {
		if p.PromptPreamble == "" {
			t.Errorf("%s has no prompt preamble", p.Name)
		}
	}
}

func TestArchetypes_AllHaveColorPref(t *testing.T) {
	for _, p := range All() {
		if p.ColorPref.Shade == "" {
			t.Errorf("%s has no shade preference", p.Name)
		}
		if p.ColorPref.Color == "" {
			t.Errorf("%s has no color preference", p.Name)
		}
	}
}

func TestArchetypes_HomeZoneMatchesPosition(t *testing.T) {
	for _, p := range All() {
		expected := HomeZoneFor(p.Position)
		if p.HomeZone != expected {
			t.Errorf("%s: HomeZone=%s, want %s (for position %s)", p.Name, p.HomeZone, expected, p.Position)
		}
	}
}

func TestArchetypes_ThreeAxes(t *testing.T) {
	for _, p := range All() {
		if p.Element == "" {
			t.Errorf("%s missing phenotype (Element)", p.Name)
		}
		if len(p.PersonalityTags) == 0 {
			t.Errorf("%s missing trait (PersonalityTags)", p.Name)
		}
		if p.Description == "" {
			t.Errorf("%s missing mission (Description)", p.Name)
		}
	}
}

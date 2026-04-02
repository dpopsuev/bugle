package identity

import (
	"context"
	"math"
	"testing"
)

func TestHeuristicInfer_FastCoding(t *testing.T) {
	vec, conf := heuristicInfer("I need a fast coding agent")
	if vec.Speed < 0.8 {
		t.Errorf("speed = %f, want >= 0.8", vec.Speed)
	}
	if vec.Coding < 0.8 {
		t.Errorf("coding = %f, want >= 0.8", vec.Coding)
	}
	if conf == 0 {
		t.Error("confidence should be > 0")
	}
	t.Logf("fast coding: speed=%.1f coding=%.1f conf=%.2f", vec.Speed, vec.Coding, conf)
}

func TestHeuristicInfer_RigorousAudit(t *testing.T) {
	vec, _ := heuristicInfer("Perform a rigorous security audit with evidence")
	if vec.Rigor < 0.8 {
		t.Errorf("rigor = %f, want >= 0.8", vec.Rigor)
	}
}

func TestHeuristicInfer_DebateCreative(t *testing.T) {
	vec, _ := heuristicInfer("Brainstorm and debate architecture options")
	if vec.Discourse < 0.7 {
		t.Errorf("discourse = %f, want >= 0.7", vec.Discourse)
	}
}

func TestHeuristicInfer_NoKeywords(t *testing.T) {
	vec, conf := heuristicInfer("Do something")
	if conf > 0 {
		t.Errorf("confidence = %f, want 0 for no keywords", conf)
	}
	// All traits should be zero.
	if vec.Speed != 0 || vec.Coding != 0 {
		t.Error("all traits should be zero for unknown intent")
	}
}

func TestHeuristicInfer_MultiTrait(t *testing.T) {
	vec, conf := heuristicInfer("Debug and investigate root cause with shell tools")
	if vec.Reasoning < 0.8 {
		t.Errorf("reasoning = %f, want >= 0.8 (investigate + root cause)", vec.Reasoning)
	}
	if vec.Coding < 0.6 {
		t.Errorf("coding = %f, want >= 0.6 (debug)", vec.Coding)
	}
	if vec.ToolUse < 0.7 {
		t.Errorf("tooluse = %f, want >= 0.7 (shell)", vec.ToolUse)
	}
	t.Logf("multi-trait: reasoning=%.1f coding=%.1f tooluse=%.1f conf=%.2f",
		vec.Reasoning, vec.Coding, vec.ToolUse, conf)
}

func TestInferFromIntent_HeuristicOnly(t *testing.T) {
	vec, err := InferFromIntent(context.Background(), "fast code review", InferConfig{})
	if err != nil {
		t.Fatalf("InferFromIntent: %v", err)
	}
	if vec.Speed < 0.8 {
		t.Errorf("speed = %f, want >= 0.8", vec.Speed)
	}
	if vec.Coding < 0.6 {
		t.Errorf("coding = %f, want >= 0.6 (review → coding)", vec.Coding)
	}
}

func TestExtractJSON(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`{"speed": 0.9}`, `{"speed": 0.9}`},
		{"Here is the JSON:\n```json\n{\"speed\": 0.9}\n```", `{"speed": 0.9}`},
		{`no json here`, `no json here`},
	}
	for _, tt := range tests {
		got := extractJSON(tt.input)
		if got != tt.want {
			t.Errorf("extractJSON(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestMergeVectors(t *testing.T) {
	heuristic := TraitVector{Speed: 0.9, Coding: 0.0}
	agentVec := TraitVector{Speed: 0.0, Coding: 0.8, Reasoning: 0.7}

	merged := mergeVectors(heuristic, agentVec, 0.3) // low heuristic confidence

	// Speed: heuristic had it, agent didn't → keep heuristic
	if merged.Speed != 0.9 {
		t.Errorf("speed = %f, want 0.9 (heuristic only)", merged.Speed)
	}
	// Coding: heuristic was 0, agent had 0.8 → use agent
	if merged.Coding != 0.8 {
		t.Errorf("coding = %f, want 0.8 (agent only)", merged.Coding)
	}
	// Reasoning: heuristic was 0, agent had 0.7 → use agent
	if merged.Reasoning != 0.7 {
		t.Errorf("reasoning = %f, want 0.7 (agent only)", merged.Reasoning)
	}
}

func TestMergeVectors_BothHaveValues(t *testing.T) {
	heuristic := TraitVector{Speed: 0.9}
	agentVec := TraitVector{Speed: 0.5}

	merged := mergeVectors(heuristic, agentVec, 0.3)

	// Both have speed — weighted: 0.9*0.3 + 0.5*0.7 = 0.27 + 0.35 = 0.62
	want := 0.9*0.3 + 0.5*0.7
	if math.Abs(merged.Speed-want) > 0.01 {
		t.Errorf("speed = %f, want %f (weighted merge)", merged.Speed, want)
	}
}

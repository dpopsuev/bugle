package identity

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Responder sends a prompt and returns a response. Used for agent-based
// trait inference when heuristics are insufficient. ISP: one method.
type Responder interface {
	RespondTo(ctx context.Context, prompt string) (string, error)
}

// InferConfig controls how intent-to-trait inference works.
type InferConfig struct {
	// Agent is an optional Responder used for ambiguous intents.
	// If nil, only heuristic matching is used.
	Agent Responder

	// Threshold is the minimum heuristic confidence (0.0-1.0) to skip
	// the agent call. Default: 0.6 (if >=60% of traits matched by keywords,
	// skip agent).
	Threshold float64
}

// InferFromIntent maps a natural language intent string to a TraitVector.
// Uses heuristic keyword matching first. If confidence is below threshold
// and an Agent is configured, delegates to the agent for refinement.
func InferFromIntent(ctx context.Context, intent string, cfg InferConfig) (TraitVector, error) {
	if cfg.Threshold == 0 {
		cfg.Threshold = 0.6
	}

	// Heuristic pass — keyword matching, $0, instant.
	vec, confidence := heuristicInfer(intent)

	if confidence >= cfg.Threshold || cfg.Agent == nil {
		return vec, nil
	}

	// Agent pass — Responder refines the heuristic result.
	agentVec, err := agentInfer(ctx, cfg.Agent, intent)
	if err != nil {
		// Agent failed — fall back to heuristic.
		return vec, nil //nolint:nilerr // graceful degradation
	}

	// Merge: agent overrides zero heuristic values, keeps strong ones.
	return mergeVectors(vec, agentVec, confidence), nil
}

// heuristicInfer extracts traits from keywords. Returns the vector and a
// confidence score (0.0-1.0) based on how many traits were matched.
func heuristicInfer(intent string) (vec TraitVector, confidence float64) {
	lower := strings.ToLower(intent)
	var matched int

	for keyword, apply := range keywordMap {
		if strings.Contains(lower, keyword) {
			apply(&vec)
			matched++
		}
	}

	// Confidence = proportion of traits matched (max 8 keywords → 1.0).
	confidence = float64(matched) / 8.0
	if confidence > 1.0 {
		confidence = 1.0
	}
	return
}

// keywordMap maps keywords to trait adjustments.
var keywordMap = map[string]func(*TraitVector){
	// Speed
	"fast":    func(v *TraitVector) { v.Speed = 0.9 },
	"quick":   func(v *TraitVector) { v.Speed = 0.9 },
	"cheap":   func(v *TraitVector) { v.Speed = 0.8 },
	"instant": func(v *TraitVector) { v.Speed = 1.0 },

	// Reasoning
	"reason":      func(v *TraitVector) { v.Reasoning = 0.9 },
	"investigate": func(v *TraitVector) { v.Reasoning = 0.8 },
	"analyze":     func(v *TraitVector) { v.Reasoning = 0.8 },
	"root cause":  func(v *TraitVector) { v.Reasoning = 0.9 },
	"debug":       func(v *TraitVector) { v.Reasoning = 0.8; v.Coding = 0.7 },

	// Rigor
	"rigorous": func(v *TraitVector) { v.Rigor = 0.9 },
	"evidence": func(v *TraitVector) { v.Rigor = 0.8 },
	"proof":    func(v *TraitVector) { v.Rigor = 0.9 },
	"verify":   func(v *TraitVector) { v.Rigor = 0.8 },
	"audit":    func(v *TraitVector) { v.Rigor = 0.9 },
	"review":   func(v *TraitVector) { v.Rigor = 0.7; v.Coding = 0.6 },

	// Coding
	"code":      func(v *TraitVector) { v.Coding = 0.9 },
	"coding":    func(v *TraitVector) { v.Coding = 0.9 },
	"implement": func(v *TraitVector) { v.Coding = 0.9 },
	"refactor":  func(v *TraitVector) { v.Coding = 0.8 },
	"fix":       func(v *TraitVector) { v.Coding = 0.8 },
	"build":     func(v *TraitVector) { v.Coding = 0.8 },
	"test":      func(v *TraitVector) { v.Coding = 0.7; v.Discipline = 0.6 },

	// Discipline
	"precise":   func(v *TraitVector) { v.Discipline = 0.9 },
	"exact":     func(v *TraitVector) { v.Discipline = 0.9 },
	"compliant": func(v *TraitVector) { v.Discipline = 0.9 },
	"follow":    func(v *TraitVector) { v.Discipline = 0.8 },

	// ToolUse
	"tool":       func(v *TraitVector) { v.ToolUse = 0.9 },
	"agentic":    func(v *TraitVector) { v.ToolUse = 0.9 },
	"autonomous": func(v *TraitVector) { v.ToolUse = 0.8 },
	"shell":      func(v *TraitVector) { v.ToolUse = 0.8; v.Coding = 0.6 },

	// Discourse
	"debate":     func(v *TraitVector) { v.Discourse = 0.9 },
	"brainstorm": func(v *TraitVector) { v.Discourse = 0.9 },
	"challenge":  func(v *TraitVector) { v.Discourse = 0.8 },
	"creative":   func(v *TraitVector) { v.Discourse = 0.8 },
	"discuss":    func(v *TraitVector) { v.Discourse = 0.7 },

	// Visual
	"visual":     func(v *TraitVector) { v.Visual = 0.9 },
	"diagram":    func(v *TraitVector) { v.Visual = 0.9 },
	"screenshot": func(v *TraitVector) { v.Visual = 0.8 },
	"image":      func(v *TraitVector) { v.Visual = 0.8 },
	"ascii":      func(v *TraitVector) { v.Visual = 0.7 },
}

const agentSystemPrompt = `You are a trait inferrer. Given a mission intent, output a JSON object with trait weights (0.0-1.0).

Available traits:
- speed: fast scan vs slow thorough analysis
- reasoning: multi-step logical chains, root cause analysis
- rigor: demands evidence, rejects uncertainty, audit-grade
- coding: read, write, debug, refactor code
- discipline: follows instructions exactly, compliance
- tooluse: chains tool calls autonomously, agentic workflows
- discourse: pushes back, challenges assumptions, brainstorms
- visual: reads screenshots, creates diagrams, ASCII art

Output ONLY valid JSON: {"speed":0.5,"reasoning":0.8,"rigor":0.3,"coding":0.9,"discipline":0.4,"tooluse":0.7,"discourse":0.2,"visual":0.1}

Omit traits that score 0. Higher = more important for this mission.`

// agentInfer asks a Responder to map intent to traits.
func agentInfer(ctx context.Context, agent Responder, intent string) (TraitVector, error) {
	prompt := fmt.Sprintf("%s\n\nMission intent: %s", agentSystemPrompt, intent)

	resp, err := agent.RespondTo(ctx, prompt)
	if err != nil {
		return TraitVector{}, fmt.Errorf("trait inference agent: %w", err)
	}

	// Extract JSON from response (agent might wrap it in markdown).
	jsonStr := extractJSON(resp)

	var vec TraitVector
	if err := json.Unmarshal([]byte(jsonStr), &vec); err != nil {
		return TraitVector{}, fmt.Errorf("trait inference parse: %w", err)
	}
	return vec, nil
}

// extractJSON finds the first JSON object in a string.
func extractJSON(s string) string {
	start := strings.Index(s, "{")
	if start < 0 {
		return s
	}
	end := strings.LastIndex(s, "}")
	if end < start {
		return s
	}
	return s[start : end+1]
}

// mergeVectors combines heuristic and agent vectors. Agent values override
// heuristic zeros. For traits where both have values, weight by confidence.
func mergeVectors(heuristic, agentVec TraitVector, heuristicConf float64) TraitVector {
	merge := func(h, a float64) float64 {
		if h == 0 {
			return a // heuristic had nothing, use agent
		}
		if a == 0 {
			return h // agent had nothing, keep heuristic
		}
		// Both have values — weighted average biased toward agent
		// (agent was called because heuristic confidence was low).
		return h*heuristicConf + a*(1-heuristicConf)
	}

	return TraitVector{
		Speed:      merge(heuristic.Speed, agentVec.Speed),
		Reasoning:  merge(heuristic.Reasoning, agentVec.Reasoning),
		Rigor:      merge(heuristic.Rigor, agentVec.Rigor),
		Coding:     merge(heuristic.Coding, agentVec.Coding),
		Discipline: merge(heuristic.Discipline, agentVec.Discipline),
		ToolUse:    merge(heuristic.ToolUse, agentVec.ToolUse),
		Discourse:  merge(heuristic.Discourse, agentVec.Discourse),
		Visual:     merge(heuristic.Visual, agentVec.Visual),
	}
}

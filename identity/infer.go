package identity

// FromVector converts an Arsenal TraitVector to a trait.Set ECS component.
// This is the v0.2.0 bridge between the catalog (predicted traits) and the
// ECS world (attached traits). Full intent-based inference is deferred.
func FromVector(v TraitVector) Set {
	return Set{
		{Name: Speed, Value: v.Speed},
		{Name: Reasoning, Value: v.Reasoning},
		{Name: Rigor, Value: v.Rigor},
		{Name: Coding, Value: v.Coding},
		{Name: Discipline, Value: v.Discipline},
		{Name: ToolUse, Value: v.ToolUse},
		{Name: Discourse, Value: v.Discourse},
		{Name: Visual, Value: v.Visual},
	}
}

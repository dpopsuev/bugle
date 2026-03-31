package arsenal

// TraitMapping defines how raw benchmark scores map to the 8 traits.
// Each trait has a map of benchmark_name → weight (weights should sum to 1.0).
type TraitMapping struct {
	Speed      map[string]float64 `yaml:"speed"`
	Reasoning  map[string]float64 `yaml:"reasoning"`
	Rigor      map[string]float64 `yaml:"rigor"`
	Coding     map[string]float64 `yaml:"coding"`
	Discipline map[string]float64 `yaml:"discipline"`
	ToolUse    map[string]float64 `yaml:"tooluse"`
	Discourse  map[string]float64 `yaml:"discourse"`
	Visual     map[string]float64 `yaml:"visual"`
}

// ApplyMapping converts raw benchmark scores to a TraitVector using weighted sums.
func ApplyMapping(benchmarks map[string]float64, m TraitMapping) TraitVector {
	return TraitVector{
		Speed:      weightedSum(benchmarks, m.Speed),
		Reasoning:  weightedSum(benchmarks, m.Reasoning),
		Rigor:      weightedSum(benchmarks, m.Rigor),
		Coding:     weightedSum(benchmarks, m.Coding),
		Discipline: weightedSum(benchmarks, m.Discipline),
		ToolUse:    weightedSum(benchmarks, m.ToolUse),
		Discourse:  weightedSum(benchmarks, m.Discourse),
		Visual:     weightedSum(benchmarks, m.Visual),
	}
}

func weightedSum(benchmarks, weights map[string]float64) float64 {
	var sum float64
	for name, weight := range weights {
		sum += benchmarks[name] * weight
	}
	return sum
}

package troupe

import "context"

// Pick selects a subset from candidates. Pure decision — no side effects.
// Returns a filtered slice (may be empty). The input slice must not be modified.
type Pick[T any] func(ctx context.Context, candidates []T) []T

// PickAll returns all candidates unchanged.
func PickAll[T any]() Pick[T] {
	return func(_ context.Context, candidates []T) []T {
		return candidates
	}
}

// PickFirst returns at most the first n candidates.
func PickFirst[T any](n int) Pick[T] {
	return func(_ context.Context, candidates []T) []T {
		if n <= 0 {
			n = 1
		}
		if n > len(candidates) {
			n = len(candidates)
		}
		return candidates[:n]
	}
}

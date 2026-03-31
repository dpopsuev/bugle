// gate_budget.go — BudgetGatekeeper: code-only gate for token budget enforcement.
//
// Proves that Gatekeeper isn't always an LLM agent — it's a contract.
// BudgetGatekeeper checks remaining budget before allowing entry.
package collective

import (
	"context"
	"fmt"
)

// BudgetGatekeeper implements Gatekeeper by checking token budget.
// A code-only gate — no LLM, no agent, just math.
type BudgetGatekeeper struct {
	MaxTokens int        // 0 = unlimited (always pass)
	Spent     func() int // callback to check current spend
}

// Pass checks if the budget allows entry.
func (g *BudgetGatekeeper) Pass(_ context.Context, _ string) (allowed bool, reason string, err error) {
	if g.MaxTokens <= 0 {
		return true, "", nil // unlimited
	}
	spent := 0
	if g.Spent != nil {
		spent = g.Spent()
	}
	if spent >= g.MaxTokens {
		return false, fmt.Sprintf("budget exceeded: %d/%d tokens", spent, g.MaxTokens), nil
	}
	return true, "", nil
}

// Compile-time check.
var _ Gatekeeper = (*BudgetGatekeeper)(nil)

package billing

import (
	"sync"
	"time"
)

// PeriodBudget enforces spending limits that reset on a time period.
// Lazy reset: checks time on every Check() call, no goroutine needed.
type PeriodBudget struct {
	limit   float64
	period  time.Duration
	spent   float64
	resetAt time.Time
	mu      sync.Mutex
}

// NewPeriodBudget creates a budget with the given limit and reset period.
func NewPeriodBudget(limit float64, period time.Duration) *PeriodBudget {
	return &PeriodBudget{
		limit:   limit,
		period:  period,
		resetAt: time.Now().Add(period),
	}
}

// Check verifies the budget has not been exceeded. Resets if the period has elapsed.
func (b *PeriodBudget) Check() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if time.Now().After(b.resetAt) {
		b.spent = 0
		b.resetAt = time.Now().Add(b.period)
	}
	if b.spent >= b.limit {
		return ErrBudgetExceeded
	}
	return nil
}

// Record adds cost to the current period's spending.
func (b *PeriodBudget) Record(cost float64) {
	b.mu.Lock()
	b.spent += cost
	b.mu.Unlock()
}

// Spent returns the current period's total spending.
func (b *PeriodBudget) Spent() float64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.spent
}

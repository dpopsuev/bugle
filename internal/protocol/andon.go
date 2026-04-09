package protocol

import "github.com/dpopsuev/troupe/signal"

// Type aliases — internal consumers continue to work unchanged.
// Andon types now live in signal/ (merged from andon/).
type (
	AndonLevel    = signal.AndonLevel
	AndonCategory = signal.Category
	Andon         = signal.Andon
	AndonLevelDef = signal.LevelDef
)

// Re-export constants.
const (
	AndonNominal  = signal.Nominal
	AndonDegraded = signal.Degraded
	AndonFailure  = signal.Failure
	AndonBlocked  = signal.Blocked
	AndonDead     = signal.Dead

	PriorityNominal  = signal.PriorityNominal
	PriorityDegraded = signal.PriorityDegraded
	PriorityFailure  = signal.PriorityFailure
	PriorityBlocked  = signal.PriorityBlocked
	PriorityDead     = signal.PriorityDead

	CategoryBudget    = signal.CategoryBudget
	CategoryDeadlock  = signal.CategoryDeadlock
	CategoryLifecycle = signal.CategoryLifecycle
	CategoryQuality   = signal.CategoryQuality
	CategorySecurity  = signal.CategorySecurity
	CategoryDrift     = signal.CategoryDrift
)

// Re-export functions.
var (
	DefaultVocabulary = signal.DefaultVocabulary
	Worse             = signal.Worse
	WorstPriority     = signal.WorstPriority
	PriorityOf        = signal.PriorityOf
	ReservedLevels    = signal.ReservedLevels
	ReservedColors    = signal.ReservedColors
)

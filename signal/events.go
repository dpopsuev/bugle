package signal

// Control bus events — routing decisions, policy enforcement.
// Emitted by: broker, director.
const (
	EventDispatchRouted = "dispatch_routed"
	EventHookExecuted   = "hook_executed"
	EventVetoApplied    = "veto_applied"
)

// Work bus events — task lifecycle.
// Emitted by: hookedActor (Perform), transport handlers.
const (
	EventWorkerStart = "start"
	EventWorkerDone  = "done"
	EventWorkerError = "error"
)

// Status bus events — observability, health, lifecycle.
// Emitted by: warden, world, supervisor, native telemetry.
const (
	EventWorkerStarted = "worker_started"
	EventWorkerStopped = "worker_stopped"
	EventShouldStop    = "should_stop"
	EventBudgetUpdate  = "budget_update"
	EventZoneShift     = "zone_shift"
)

// Signal meta key constants used in bus.Emit meta maps and read by
// Supervisor.Process and other signal consumers.
const (
	MetaKeyWorkerID       = "worker_id"
	MetaKeyError          = "error"
	MetaKeyUsed           = "used"
	MetaKeyFromZone       = "from_zone"
	MetaKeyToZone         = "to_zone"
	MetaKeyMode           = "mode"
	MetaKeyBytes          = "bytes"
	MetaKeyInFlight       = "in_flight"
	MetaKeyVia            = "via"
	MetaKeyPromptPath     = "prompt_path"
	MetaKeyDispatchReason = "dispatch_reason"
	MetaKeyQueueDepth     = "queue_depth"
	MetaKeyHookName       = "hook_name"
	MetaKeyHookPhase      = "hook_phase"
	MetaKeyShade          = "shade"
	MetaKeyColor          = "color"
)

// Agent name constants used as the agent field in Signal.
const (
	AgentWorker     = "worker"
	AgentSupervisor = "supervisor"
	AgentServer     = "server"
	AgentMediator   = "mediator"
)

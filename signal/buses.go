package signal

// ControlLog carries control-plane events: routing decisions, vetoes,
// hook executions. Producers: broker, director.
type ControlLog struct{ EventLog }

// WorkLog carries data-plane events: task start, done, error.
// Producers: hookedActor, transport handlers.
type WorkLog struct{ EventLog }

// StatusLog carries observability events: Andon health transitions,
// worker lifecycle, budget updates, perf metrics, and projections
// from Control and Data planes. Producers: warden, world, supervisor,
// plus projected events from control/data emitters.
type StatusLog struct{ EventLog }

// BusSet groups the three typed event buses. Passed through broker,
// warden, and testkit instead of a single EventLog.
type BusSet struct {
	Control ControlLog
	Work    WorkLog
	Status  StatusLog
}

// NewBusSet creates a BusSet backed by three independent MemLogs.
func NewBusSet() BusSet {
	return BusSet{
		Control: ControlLog{NewMemLog()},
		Work:    WorkLog{NewMemLog()},
		Status:  StatusLog{NewMemLog()},
	}
}

// DurableBusSet wraps three DurableEventLogs for persistent bus storage.
type DurableBusSet struct {
	BusSet
	control *DurableEventLog
	work    *DurableEventLog
	status  *DurableEventLog
}

// NewDurableBusSet creates a DurableBusSet backed by JSON-Lines files
// in the given directory: control.jsonl, work.jsonl, status.jsonl.
func NewDurableBusSet(dir string) (*DurableBusSet, error) {
	control, err := NewDurableJSONLines(dir + "/control.jsonl")
	if err != nil {
		return nil, err
	}
	work, err := NewDurableJSONLines(dir + "/work.jsonl")
	if err != nil {
		control.Close()
		return nil, err
	}
	status, err := NewDurableJSONLines(dir + "/status.jsonl")
	if err != nil {
		control.Close()
		work.Close()
		return nil, err
	}
	return &DurableBusSet{
		BusSet: BusSet{
			Control: ControlLog{control},
			Work:    WorkLog{work},
			Status:  StatusLog{status},
		},
		control: control,
		work:    work,
		status:  status,
	}, nil
}

func (d *DurableBusSet) logs() []*DurableEventLog {
	return []*DurableEventLog{d.control, d.work, d.status}
}

// Replay reads persisted events from all three stores into memory.
func (d *DurableBusSet) Replay() (int, error) {
	total := 0
	for _, log := range d.logs() {
		n, err := log.Replay()
		if err != nil {
			return total, err
		}
		total += n
	}
	return total, nil
}

// Close flushes and closes all three stores.
func (d *DurableBusSet) Close() error {
	var firstErr error
	for _, log := range d.logs() {
		if err := log.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

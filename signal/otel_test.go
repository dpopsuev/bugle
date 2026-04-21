package signal_test

import (
	"context"
	"testing"

	"github.com/dpopsuev/troupe/signal"
)

func TestOTelAdapter_SubscribesAndEmits(t *testing.T) {
	buses := signal.NewBusSet()

	adapter, err := signal.NewOTelAdapter(context.Background(), "test-service")
	if err != nil {
		t.Fatalf("NewOTelAdapter: %v", err)
	}
	defer adapter.Close()

	adapter.Subscribe(buses)

	buses.Control.Emit(signal.Event{
		Kind:   signal.EventDispatchRouted,
		Source: "broker",
	})
	buses.Work.Emit(signal.Event{
		Kind:   signal.EventWorkerStart,
		Source: "worker-1",
	})
	buses.Work.Emit(signal.Event{
		Kind:   signal.EventWorkerError,
		Source: "worker-1",
	})
	buses.Status.Emit(signal.Event{
		Kind:   signal.EventWorkerStarted,
		Source: "warden",
	})

	if buses.Control.Len() != 1 {
		t.Errorf("control events = %d, want 1", buses.Control.Len())
	}
	if buses.Work.Len() != 2 {
		t.Errorf("work events = %d, want 2", buses.Work.Len())
	}
	if buses.Status.Len() != 1 {
		t.Errorf("status events = %d, want 1", buses.Status.Len())
	}
}

func TestOTelAdapter_CloseEndsSpan(t *testing.T) {
	adapter, err := signal.NewOTelAdapter(context.Background(), "test-close")
	if err != nil {
		t.Fatalf("NewOTelAdapter: %v", err)
	}
	adapter.Close()
}

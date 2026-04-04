package billing

import (
	"testing"
	"time"
)

func TestPeriodBudget_UnderLimit(t *testing.T) {
	pb := NewPeriodBudget(1.0, time.Hour)
	pb.Record(0.5)
	if err := pb.Check(); err != nil {
		t.Fatalf("under limit should pass: %v", err)
	}
}

func TestPeriodBudget_ExceedsLimit(t *testing.T) {
	pb := NewPeriodBudget(1.0, time.Hour)
	pb.Record(1.5)
	if err := pb.Check(); err == nil {
		t.Fatal("expected budget exceeded")
	}
}

func TestPeriodBudget_LazyReset(t *testing.T) {
	pb := NewPeriodBudget(1.0, time.Millisecond)
	pb.Record(1.5) // exceed
	time.Sleep(2 * time.Millisecond)
	if err := pb.Check(); err != nil {
		t.Fatalf("should have reset: %v", err)
	}
	if pb.Spent() != 0 {
		t.Errorf("spent = %f, want 0 after reset", pb.Spent())
	}
}

func TestPeriodBudget_ExactBoundary(t *testing.T) {
	pb := NewPeriodBudget(1.0, time.Hour)
	pb.Record(1.0)
	if err := pb.Check(); err == nil {
		t.Fatal("exact limit should be rejected (>=)")
	}
}

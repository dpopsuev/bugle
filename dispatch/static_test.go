package dispatch

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNewStaticDispatcher(t *testing.T) {
	d := NewStaticDispatcher("")
	if d == nil {
		t.Fatal("NewStaticDispatcher returned nil")
	}
}

func TestStaticDispatcher_InMemory(t *testing.T) {
	d := NewStaticDispatcher("")

	data := json.RawMessage(`{"category":"infra"}`)
	d.Set("C1", "F0_RECALL", data)

	got, err := d.Dispatch(context.Background(), Context{CaseID: "C1", Step: "F0_RECALL"})
	if err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if string(got) != string(data) {
		t.Errorf("got %s, want %s", got, data)
	}
}

func TestStaticDispatcher_MissingArtifact(t *testing.T) {
	d := NewStaticDispatcher("")
	_, err := d.Dispatch(context.Background(), Context{CaseID: "C99", Step: "F0_RECALL"})
	if err == nil {
		t.Error("expected error for missing artifact")
	}
}

func TestStaticDispatcher_FromDirectory(t *testing.T) {
	dir := t.TempDir()
	caseDir := filepath.Join(dir, "C1")
	if err := os.MkdirAll(caseDir, 0o755); err != nil {
		t.Fatal(err)
	}

	content := `{"severity":"high"}`
	if err := os.WriteFile(filepath.Join(caseDir, "F1_TRIAGE.json"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	d := NewStaticDispatcher(dir)

	got, err := d.Dispatch(context.Background(), Context{CaseID: "C1", Step: "F1_TRIAGE"})
	if err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if string(got) != content {
		t.Errorf("got %s, want %s", got, content)
	}
}

func TestStaticDispatcher_CaseInsensitiveFallback(t *testing.T) {
	dir := t.TempDir()
	caseDir := filepath.Join(dir, "C2")
	if err := os.MkdirAll(caseDir, 0o755); err != nil {
		t.Fatal(err)
	}

	content := `{"matched":"lowercase"}`
	if err := os.WriteFile(filepath.Join(caseDir, "f0_recall.json"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	d := NewStaticDispatcher(dir)

	got, err := d.Dispatch(context.Background(), Context{CaseID: "C2", Step: "F0_RECALL"})
	if err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if string(got) != content {
		t.Errorf("got %s, want %s", got, content)
	}
}

func TestStaticDispatcher_InMemoryOverridesFile(t *testing.T) {
	dir := t.TempDir()
	caseDir := filepath.Join(dir, "C1")
	if err := os.MkdirAll(caseDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(caseDir, "F0_RECALL.json"), []byte(`{"from":"file"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	d := NewStaticDispatcher(dir)
	d.Set("C1", "F0_RECALL", json.RawMessage(`{"from":"memory"}`))

	got, err := d.Dispatch(context.Background(), Context{CaseID: "C1", Step: "F0_RECALL"})
	if err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if string(got) != `{"from":"memory"}` {
		t.Errorf("in-memory should take precedence, got %s", got)
	}
}

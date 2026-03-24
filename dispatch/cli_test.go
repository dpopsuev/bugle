package dispatch

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var _ Dispatcher = (*CLIDispatcher)(nil)

func TestNewCLIDispatcher_ValidCommand(t *testing.T) {
	d, err := NewCLIDispatcher("echo")
	if err != nil {
		t.Fatalf("echo should be found: %v", err)
	}
	if d.Timeout != 5*time.Minute {
		t.Errorf("default timeout = %v, want 5m", d.Timeout)
	}
}

func TestNewCLIDispatcher_InvalidCommand(t *testing.T) {
	_, err := NewCLIDispatcher("nonexistent-binary-xyz-12345")
	if err == nil {
		t.Fatal("expected error for missing binary")
	}
	if !strings.Contains(err.Error(), "not found in PATH") {
		t.Errorf("error = %v, want 'not found in PATH'", err)
	}
}

func TestNewCLIDispatcher_Options(t *testing.T) {
	d, err := NewCLIDispatcher("echo",
		WithCLIArgs("--json"),
		WithCLITimeout(30*time.Second),
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Args) != 1 || d.Args[0] != "--json" {
		t.Errorf("Args = %v", d.Args)
	}
	if d.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, want 30s", d.Timeout)
	}
}

func TestCLIDispatcher_Dispatch_Echo(t *testing.T) {
	dir := t.TempDir()
	promptPath := filepath.Join(dir, "prompt.txt")
	artifactPath := filepath.Join(dir, "artifact.json")

	if err := os.WriteFile(promptPath, []byte("test prompt content"), 0o644); err != nil {
		t.Fatal(err)
	}

	d, err := NewCLIDispatcher("cat")
	if err != nil {
		t.Fatalf("cat should exist: %v", err)
	}

	result, err := d.Dispatch(context.Background(), &Context{
		DispatchID:   1,
		CaseID:       "C01",
		Step:         "F1_TRIAGE",
		PromptPath:   promptPath,
		ArtifactPath: artifactPath,
	})
	if err != nil {
		t.Fatalf("Dispatch: %v", err)
	}

	if string(result) != "test prompt content" {
		t.Errorf("result = %q, want %q", string(result), "test prompt content")
	}

	saved, err := os.ReadFile(artifactPath)
	if err != nil {
		t.Fatalf("read artifact: %v", err)
	}
	if string(saved) != "test prompt content" {
		t.Errorf("saved artifact = %q", string(saved))
	}
}

func TestCLIDispatcher_Dispatch_WithArgs(t *testing.T) {
	dir := t.TempDir()
	promptPath := filepath.Join(dir, "prompt.txt")
	artifactPath := filepath.Join(dir, "artifact.json")

	if err := os.WriteFile(promptPath, []byte("hello world hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	d, err := NewCLIDispatcher("tr",
		WithCLIArgs("o", "0"),
	)
	if err != nil {
		t.Fatalf("tr should exist: %v", err)
	}

	result, err := d.Dispatch(context.Background(), &Context{
		DispatchID:   2,
		CaseID:       "C02",
		Step:         "F0_RECALL",
		PromptPath:   promptPath,
		ArtifactPath: artifactPath,
	})
	if err != nil {
		t.Fatalf("Dispatch: %v", err)
	}

	if string(result) != "hell0 w0rld hell0" {
		t.Errorf("result = %q", string(result))
	}
}

func TestCLIDispatcher_Dispatch_MissingPrompt(t *testing.T) {
	d, _ := NewCLIDispatcher("cat")
	_, err := d.Dispatch(context.Background(), &Context{
		PromptPath:   "/nonexistent/prompt.txt",
		ArtifactPath: "/tmp/out.json",
	})
	if err == nil {
		t.Fatal("expected error for missing prompt")
	}
	if !strings.Contains(err.Error(), "read prompt") {
		t.Errorf("error = %v", err)
	}
}

func TestCLIDispatcher_Dispatch_CommandFailure(t *testing.T) {
	dir := t.TempDir()
	promptPath := filepath.Join(dir, "prompt.txt")
	if err := os.WriteFile(promptPath, []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}

	d, err := NewCLIDispatcher("false")
	if err != nil {
		t.Fatal(err)
	}

	_, err = d.Dispatch(context.Background(), &Context{
		PromptPath:   promptPath,
		ArtifactPath: filepath.Join(dir, "out.json"),
	})
	if err == nil {
		t.Fatal("expected error for failed command")
	}
	if !strings.Contains(err.Error(), "command failed") {
		t.Errorf("error = %v", err)
	}
}

func TestCLIDispatcher_Dispatch_Timeout(t *testing.T) {
	dir := t.TempDir()
	promptPath := filepath.Join(dir, "prompt.txt")
	if err := os.WriteFile(promptPath, []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}

	d, err := NewCLIDispatcher("sleep",
		WithCLIArgs("10"),
		WithCLITimeout(100*time.Millisecond),
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = d.Dispatch(context.Background(), &Context{
		PromptPath:   promptPath,
		ArtifactPath: filepath.Join(dir, "out.json"),
	})
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("error = %v, want timeout", err)
	}
}

func TestCLIDispatcher_Dispatch_EmptyOutput(t *testing.T) {
	dir := t.TempDir()
	promptPath := filepath.Join(dir, "prompt.txt")
	if err := os.WriteFile(promptPath, []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}

	d, err := NewCLIDispatcher("true")
	if err != nil {
		t.Fatal(err)
	}

	_, err = d.Dispatch(context.Background(), &Context{
		PromptPath:   promptPath,
		ArtifactPath: filepath.Join(dir, "out.json"),
	})
	if err == nil {
		t.Fatal("expected error for empty output")
	}
	if !strings.Contains(err.Error(), "no output") {
		t.Errorf("error = %v", err)
	}
}

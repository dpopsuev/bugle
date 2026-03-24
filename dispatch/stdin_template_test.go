package dispatch

import (
	"testing"
)

func TestDefaultStdinTemplate(t *testing.T) {
	tmpl := DefaultStdinTemplate()
	if len(tmpl.Instructions) == 0 {
		t.Fatal("default template should have instructions")
	}
	if len(tmpl.Instructions) != 3 {
		t.Errorf("expected 3 default instructions, got %d", len(tmpl.Instructions))
	}
}

func TestNewStdinDispatcher_UsesDefault(t *testing.T) {
	d := NewStdinDispatcher()
	if len(d.template.Instructions) != 3 {
		t.Errorf("NewStdinDispatcher should use default template, got %d instructions",
			len(d.template.Instructions))
	}
}

func TestNewStdinDispatcherWithTemplate_Custom(t *testing.T) {
	custom := StdinTemplate{
		Instructions: []string{
			"1. Run the vulnerability scan",
			"2. Save the report to the artifact path",
			"3. Press Enter when done",
		},
	}
	d := NewStdinDispatcherWithTemplate(custom)
	if len(d.template.Instructions) != 3 {
		t.Fatalf("expected 3 custom instructions, got %d", len(d.template.Instructions))
	}
	if d.template.Instructions[0] != "1. Run the vulnerability scan" {
		t.Errorf("first instruction mismatch: %q", d.template.Instructions[0])
	}
}

func TestNewStdinDispatcherWithTemplate_Empty(t *testing.T) {
	d := NewStdinDispatcherWithTemplate(StdinTemplate{})
	if len(d.template.Instructions) != 0 {
		t.Errorf("empty template should have 0 instructions, got %d",
			len(d.template.Instructions))
	}
}

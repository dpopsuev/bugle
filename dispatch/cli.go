package dispatch

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"time"
)

// CLIDispatcher shells out to an external CLI-based LLM tool (e.g. Codex,
// Claude CLI) to process a prompt and produce an artifact.
//
// The dispatcher reads the prompt file, pipes its content to the CLI's stdin,
// captures stdout as the artifact, and writes it to the artifact path.
//
// This is a PoC battery — sufficient for prototyping with any CLI tool
// that accepts a prompt on stdin and returns a response on stdout.
type CLIDispatcher struct {
	Command string
	Args    []string
	Timeout time.Duration
	Logger  *slog.Logger
}

// CLIOption configures a CLIDispatcher.
type CLIOption func(*CLIDispatcher)

// WithCLIArgs sets additional arguments passed to the CLI command.
func WithCLIArgs(args ...string) CLIOption {
	return func(d *CLIDispatcher) { d.Args = args }
}

// WithCLITimeout sets the maximum execution time for a single dispatch.
// Defaults to 5 minutes.
func WithCLITimeout(t time.Duration) CLIOption {
	return func(d *CLIDispatcher) { d.Timeout = t }
}

// WithCLILogger sets a structured logger.
func WithCLILogger(l *slog.Logger) CLIOption {
	return func(d *CLIDispatcher) { d.Logger = l }
}

// NewCLIDispatcher creates a dispatcher that invokes the given command for
// each circuit step. The command must accept the prompt on stdin and write
// the artifact JSON to stdout.
//
// The command path is validated at construction time; an error is returned
// if the binary cannot be found in $PATH.
func NewCLIDispatcher(command string, opts ...CLIOption) (*CLIDispatcher, error) {
	resolved, err := exec.LookPath(command)
	if err != nil {
		return nil, fmt.Errorf("dispatch/cli: command %q not found in PATH: %w", command, err)
	}

	d := &CLIDispatcher{
		Command: resolved,
		Timeout: 5 * time.Minute,
		Logger:  DiscardLogger(),
	}
	for _, o := range opts {
		o(d)
	}
	return d, nil
}

// Dispatch reads the prompt from PromptPath, pipes it to the CLI command's
// stdin, captures stdout as the artifact, and writes it to ArtifactPath.
func (d *CLIDispatcher) Dispatch(ctx context.Context, dctx Context) ([]byte, error) { //nolint:gocritic // value receiver for API compat
	prompt, err := os.ReadFile(dctx.PromptPath)
	if err != nil {
		return nil, fmt.Errorf("dispatch/cli: read prompt: %w", err)
	}

	cmdCtx, cancel := context.WithTimeout(ctx, d.Timeout)
	defer cancel()

	args := make([]string, len(d.Args))
	copy(args, d.Args)

	cmd := exec.CommandContext(cmdCtx, d.Command, args...)
	cmd.Stdin = bytes.NewReader(prompt)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	d.Logger.Info("dispatching CLI command",
		slog.String("command", d.Command),
		slog.String("case_id", dctx.CaseID),
		slog.String("step", dctx.Step),
	)

	start := time.Now()
	if err := cmd.Run(); err != nil {
		stderrStr := stderr.String()
		if cmdCtx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("dispatch/cli: command timed out after %v (stderr: %s)", d.Timeout, stderrStr)
		}
		return nil, fmt.Errorf("dispatch/cli: command failed: %w (stderr: %s)", err, stderrStr)
	}

	output := stdout.Bytes()
	elapsed := time.Since(start)

	if len(output) == 0 {
		return nil, fmt.Errorf("dispatch/cli: command produced no output (stderr: %s)", stderr.String())
	}

	if err := os.WriteFile(dctx.ArtifactPath, output, 0o644); err != nil {
		return nil, fmt.Errorf("dispatch/cli: write artifact: %w", err)
	}

	d.Logger.Info("CLI dispatch complete",
		slog.String("case_id", dctx.CaseID),
		slog.String("step", dctx.Step),
		slog.Int("response_bytes", len(output)),
		slog.Duration("elapsed", elapsed),
	)

	return output, nil
}

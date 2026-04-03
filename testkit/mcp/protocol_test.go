package mcp

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/dpopsuev/jericho/work"
)

// runLoop simulates the protocol loop without MCP transport.
// This is the testable core of orchestrate.RunWorker.
//
//nolint:unparam // sessionID varies in production, fixed in tests for simplicity
func runLoop(ctx context.Context, server work.Server, responder work.Responder, sessionID, workerID string, andonFn func() *work.Andon, budgetFn func() *work.BudgetActual) error {
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		pullResp, err := server.Pull(ctx, work.PullRequest{
			Action:    work.ActionPull,
			SessionID: sessionID,
			WorkerID:  workerID,
		})
		if err != nil {
			return err
		}

		if pullResp.Andon == work.AndonDead {
			return nil
		}
		if pullResp.Done {
			return nil
		}
		if !pullResp.Available {
			continue
		}

		response, err := responder.RespondTo(ctx, pullResp.PromptContent)

		pushReq := work.PushRequest{
			Action:     work.ActionPush,
			SessionID:  sessionID,
			WorkerID:   workerID,
			DispatchID: pullResp.DispatchID,
			Item:       pullResp.Item,
		}

		if err != nil {
			pushReq.Status = work.StatusBlocked
			pushReq.Fields = []byte(`{"reason":"` + err.Error() + `"}`)
		} else {
			pushReq.Status = work.StatusOk
			pushReq.Fields = []byte(response)
		}

		if andonFn != nil {
			pushReq.Andon = andonFn()
		}
		if budgetFn != nil {
			pushReq.Budget = budgetFn()
		}

		if _, pushErr := server.Push(ctx, pushReq); pushErr != nil {
			return pushErr
		}
	}
}

func TestProtocol_WorkerAbortsOnAndonDead(t *testing.T) {
	server := NewMockServer()
	server.OnPull(func(_ work.PullRequest) (work.PullResponse, error) {
		return work.PullResponse{Andon: work.AndonDead}, nil
	})

	err := runLoop(context.Background(), server, &StaticResponder{Response: "x"}, "s1", "w1", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	AssertNoPushes(t, server)
}

func TestProtocol_WorkerPushesBlockedOnResponderFailure(t *testing.T) {
	var pullCount atomic.Int32
	server := NewMockServer()
	server.OnPull(func(_ work.PullRequest) (work.PullResponse, error) {
		n := pullCount.Add(1)
		if n == 1 {
			return work.PullResponse{Available: true, Item: "F0", DispatchID: 1, PromptContent: "test"}, nil
		}
		return work.PullResponse{Done: true}, nil
	})

	responder := &FailingResponder{Err: errors.New("agent crashed")}

	err := runLoop(context.Background(), server, responder, "s1", "w1", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	AssertPushCount(t, server, 1)
	AssertPushStatus(t, server, work.StatusBlocked)
}

func TestProtocol_BudgetActualIncludedWhenFuncSet(t *testing.T) {
	var pullCount atomic.Int32
	server := NewMockServer()
	server.OnPull(func(_ work.PullRequest) (work.PullResponse, error) {
		n := pullCount.Add(1)
		if n == 1 {
			return work.PullResponse{Available: true, Item: "F0", DispatchID: 1, PromptContent: "test"}, nil
		}
		return work.PullResponse{Done: true}, nil
	})

	budgetFn := func() *work.BudgetActual {
		return &work.BudgetActual{TokensIn: 500, TokensOut: 300}
	}

	err := runLoop(context.Background(), server, &StaticResponder{Response: `{"ok":true}`}, "s1", "w1", nil, budgetFn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	AssertPushCount(t, server, 1)
	AssertBudgetReported(t, server)
}

func TestProtocol_AndonIncludedWhenFuncSet(t *testing.T) {
	var pullCount atomic.Int32
	server := NewMockServer()
	server.OnPull(func(_ work.PullRequest) (work.PullResponse, error) {
		n := pullCount.Add(1)
		if n == 1 {
			return work.PullResponse{Available: true, Item: "F0", DispatchID: 1, PromptContent: "test"}, nil
		}
		return work.PullResponse{Done: true}, nil
	})

	andonFn := func() *work.Andon {
		return &work.Andon{Level: work.AndonDegraded, Priority: work.PriorityDegraded, Message: "82% tokens"}
	}

	err := runLoop(context.Background(), server, &StaticResponder{Response: `{"ok":true}`}, "s1", "w1", andonFn, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	AssertPushCount(t, server, 1)
	AssertAndonLevel(t, server, work.AndonDegraded)
}

func TestProtocol_WorkerIDSentOnEveryPush(t *testing.T) {
	var pullCount atomic.Int32
	server := NewMockServer()
	server.OnPull(func(_ work.PullRequest) (work.PullResponse, error) {
		n := pullCount.Add(1)
		if n <= 3 {
			return work.PullResponse{Available: true, Item: "F0", DispatchID: int64(n), PromptContent: "test"}, nil
		}
		return work.PullResponse{Done: true}, nil
	})

	err := runLoop(context.Background(), server, &StaticResponder{Response: `{}`}, "s1", "[Azure·Cerulean|Analyst]", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	AssertPushCount(t, server, 3)
	AssertWorkerID(t, server, "[Azure·Cerulean|Analyst]")
}

func TestProtocol_MultipleItemsProcessedSequentially(t *testing.T) {
	var pullCount atomic.Int32
	server := NewMockServer()
	server.OnPull(func(_ work.PullRequest) (work.PullResponse, error) {
		n := pullCount.Add(1)
		if n <= 5 {
			return work.PullResponse{Available: true, Item: "item", DispatchID: int64(n), PromptContent: "test"}, nil
		}
		return work.PullResponse{Done: true}, nil
	})

	scripted := NewScriptedResponder("r1", "r2", "r3", "r4", "r5")
	err := runLoop(context.Background(), server, scripted, "s1", "w1", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	AssertPushCount(t, server, 5)
	if scripted.CallCount() != 5 {
		t.Errorf("ScriptedResponder calls = %d, want 5", scripted.CallCount())
	}
}

func TestProtocol_DoneSignalStopsLoop(t *testing.T) {
	server := NewMockServer()
	server.OnPull(func(_ work.PullRequest) (work.PullResponse, error) {
		return work.PullResponse{Done: true}, nil
	})

	err := runLoop(context.Background(), server, &StaticResponder{Response: "x"}, "s1", "w1", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	AssertNoPushes(t, server)
}

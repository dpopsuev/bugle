package transport

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/a2aproject/a2a-go/a2a"
	"github.com/a2aproject/a2a-go/a2asrv"
	"github.com/a2aproject/a2a-go/a2asrv/eventqueue"
)

// troupeExecutor implements a2asrv.AgentExecutor by dispatching to the
// baseTransport's registered handlers. This bridges A2A JSON-RPC → Troupe.
type troupeExecutor struct {
	transport *baseTransport
}

func (e *troupeExecutor) Execute(ctx context.Context, reqCtx *a2asrv.RequestContext, queue eventqueue.Queue) error {
	msg := reqCtx.Message
	slog.InfoContext(ctx, "a2a execute",
		slog.String("task_id", string(reqCtx.TaskID)),
		slog.String("role", string(msg.Role)),
		slog.Int("parts", len(msg.Parts)),
	)
	internalMsg := FromA2AMessage(*msg, AgentID(string(reqCtx.TaskID)))

	submitted := a2a.NewStatusUpdateEvent(reqCtx, a2a.TaskStateSubmitted, nil)
	if err := queue.Write(ctx, submitted); err != nil {
		return fmt.Errorf("write submitted: %w", err)
	}

	working := a2a.NewStatusUpdateEvent(reqCtx, a2a.TaskStateWorking, nil)
	if err := queue.Write(ctx, working); err != nil {
		return fmt.Errorf("write working: %w", err)
	}

	e.transport.mu.RLock()
	var handler MsgHandler
	for _, h := range e.transport.handlers {
		handler = h
		break
	}
	e.transport.mu.RUnlock()

	if handler == nil {
		slog.WarnContext(ctx, "a2a no handler registered",
			slog.String("task_id", string(reqCtx.TaskID)),
		)
		failed := a2a.NewStatusUpdateEvent(reqCtx, a2a.TaskStateFailed,
			a2a.NewMessage(a2a.MessageRoleAgent, &a2a.TextPart{Text: "no handler registered"}))
		failed.Final = true
		return queue.Write(ctx, failed)
	}

	slog.InfoContext(ctx, "a2a dispatching to handler",
		slog.String("task_id", string(reqCtx.TaskID)),
		slog.String("content", internalMsg.Content),
	)

	resp, err := handler(ctx, internalMsg)
	if err != nil {
		slog.WarnContext(ctx, "a2a handler error",
			slog.String("task_id", string(reqCtx.TaskID)),
			slog.String("error", err.Error()),
		)
		failed := a2a.NewStatusUpdateEvent(reqCtx, a2a.TaskStateFailed,
			a2a.NewMessage(a2a.MessageRoleAgent, &a2a.TextPart{Text: err.Error()}))
		failed.Final = true
		return queue.Write(ctx, failed)
	}

	slog.InfoContext(ctx, "a2a handler responded",
		slog.String("task_id", string(reqCtx.TaskID)),
		slog.String("content", resp.Content),
	)

	a2aResp := ToA2AMessage(resp)
	completed := a2a.NewStatusUpdateEvent(reqCtx, a2a.TaskStateCompleted, &a2aResp)
	completed.Final = true
	return queue.Write(ctx, completed)
}

func (e *troupeExecutor) Cancel(ctx context.Context, reqCtx *a2asrv.RequestContext, queue eventqueue.Queue) error {
	canceled := a2a.NewStatusUpdateEvent(reqCtx, a2a.TaskStateCanceled, nil)
	canceled.Final = true
	return queue.Write(ctx, canceled)
}

// A2AServerMux creates an http.ServeMux that serves A2A v1.0 JSON-RPC
// at the root path and agent card at /.well-known/agent.json.
func A2AServerMux(bt *baseTransport, card a2a.AgentCard) *http.ServeMux {
	executor := &troupeExecutor{transport: bt}
	requestHandler := a2asrv.NewHandler(executor)
	jsonrpcHandler := a2asrv.NewJSONRPCHandler(requestHandler)
	cardHandler := a2asrv.NewStaticAgentCardHandler(&card)

	mux := http.NewServeMux()
	mux.Handle("/", jsonrpcHandler)
	mux.Handle("/.well-known/agent.json", cardHandler)
	mux.Handle(a2asrv.WellKnownAgentCardPath, cardHandler)
	return mux
}

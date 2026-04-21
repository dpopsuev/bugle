package signal

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const instrumentationName = "github.com/dpopsuev/troupe/signal"

// OTelAdapter subscribes to a BusSet and emits OpenTelemetry
// traces + metrics for every signal event. Zero changes to existing
// emit points — the adapter observes via OnEmit callbacks.
type OTelAdapter struct {
	tracer  trace.Tracer
	meter   metric.Meter
	ctx     context.Context
	cancel  context.CancelFunc
	rootSpan trace.Span

	eventCount metric.Int64Counter
	errorCount metric.Int64Counter
	tokenGauge metric.Int64UpDownCounter
}

// NewOTelAdapter creates an adapter that bridges BusSet events to OTel.
// Call Close() when done to end the root span.
func NewOTelAdapter(ctx context.Context, serviceName string) (*OTelAdapter, error) {
	tracer := otel.Tracer(instrumentationName)
	meter := otel.Meter(instrumentationName)

	spanCtx, rootSpan := tracer.Start(ctx, serviceName,
		trace.WithSpanKind(trace.SpanKindInternal),
	)

	eventCount, err := meter.Int64Counter("troupe.events.total",
		metric.WithDescription("Total signal events emitted"),
	)
	if err != nil {
		rootSpan.End()
		return nil, err
	}

	errorCount, err := meter.Int64Counter("troupe.errors.total",
		metric.WithDescription("Total error events"),
	)
	if err != nil {
		rootSpan.End()
		return nil, err
	}

	tokenGauge, err := meter.Int64UpDownCounter("troupe.tokens.used",
		metric.WithDescription("Token usage across agents"),
	)
	if err != nil {
		rootSpan.End()
		return nil, err
	}

	childCtx, cancel := context.WithCancel(spanCtx)

	return &OTelAdapter{
		tracer:     tracer,
		meter:      meter,
		ctx:        childCtx,
		cancel:     cancel,
		rootSpan:   rootSpan,
		eventCount: eventCount,
		errorCount: errorCount,
		tokenGauge: tokenGauge,
	}, nil
}

// Subscribe wires the adapter to all three buses in a BusSet.
func (a *OTelAdapter) Subscribe(buses BusSet) {
	buses.Control.OnEmit(a.onControl)
	buses.Work.OnEmit(a.onWork)
	buses.Status.OnEmit(a.onStatus)
}

func (a *OTelAdapter) onControl(e Event) {
	_, span := a.tracer.Start(a.ctx, "troupe.control."+e.Kind,
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	span.SetAttributes(
		attribute.String("troupe.event.id", e.ID),
		attribute.String("troupe.event.kind", e.Kind),
		attribute.String("troupe.event.source", e.Source),
		attribute.String("troupe.bus", "control"),
	)
	if e.TraceID != "" {
		span.SetAttributes(attribute.String("troupe.trace_id", e.TraceID))
	}
	span.End()

	a.eventCount.Add(a.ctx, 1,
		metric.WithAttributes(
			attribute.String("bus", "control"),
			attribute.String("kind", e.Kind),
		),
	)
}

func (a *OTelAdapter) onWork(e Event) {
	_, span := a.tracer.Start(a.ctx, "troupe.work."+e.Kind,
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	span.SetAttributes(
		attribute.String("troupe.event.id", e.ID),
		attribute.String("troupe.event.kind", e.Kind),
		attribute.String("troupe.event.source", e.Source),
		attribute.String("troupe.bus", "work"),
	)
	if e.Kind == EventWorkerError {
		a.errorCount.Add(a.ctx, 1,
			metric.WithAttributes(attribute.String("source", e.Source)),
		)
	}
	span.End()

	a.eventCount.Add(a.ctx, 1,
		metric.WithAttributes(
			attribute.String("bus", "work"),
			attribute.String("kind", e.Kind),
		),
	)
}

func (a *OTelAdapter) onStatus(e Event) {
	_, span := a.tracer.Start(a.ctx, "troupe.status."+e.Kind,
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	span.SetAttributes(
		attribute.String("troupe.event.id", e.ID),
		attribute.String("troupe.event.kind", e.Kind),
		attribute.String("troupe.event.source", e.Source),
		attribute.String("troupe.bus", "status"),
	)
	span.End()

	a.eventCount.Add(a.ctx, 1,
		metric.WithAttributes(
			attribute.String("bus", "status"),
			attribute.String("kind", e.Kind),
		),
	)
}

// Close ends the root span and cancels the context.
func (a *OTelAdapter) Close() {
	a.cancel()
	a.rootSpan.End()
}

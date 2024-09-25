package trace

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var tracer trace.Tracer

const appTracerName = "github.com/cox96de/runner/telemetry/trace"

func init() {
	// Initialize the global trace provider. But the tracer provider is not set yet.
	// It's just to omit nil pointer dereference.
	tracer = otel.GetTracerProvider().Tracer(appTracerName)
}

// Init initializes the tracer with the global tracer provider.
func Init() {
	tracer = otel.GetTracerProvider().Tracer(appTracerName)
}

// Start starts a span with the given name and options.
func Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return tracer.Start(ctx, spanName, opts...)
}

var WithAttributes = trace.WithAttributes

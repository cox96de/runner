package main

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/cox96de/runner/log"
	"github.com/cox96de/runner/telemetry/trace"
	"go.opentelemetry.io/contrib/bridges/otellogrus"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	otellog "go.opentelemetry.io/otel/sdk/log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// setupOTEL bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func setupOTEL(ctx context.Context) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}
	defer func() {
		if err != nil {
			if shutDownErr := shutdown(ctx); shutDownErr != nil {
				err = errors.Join(err, shutDownErr)
			}
		}
	}()

	prop := newPropagator()
	otel.SetTextMapPropagator(prop)
	tracerProvider, err := newTraceProvider(ctx)
	if err != nil {
		return
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	meterProvider, err := newMeterProvider(ctx)
	if err != nil {
		return
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	loggerProvider, err := newLoggerProvider(ctx)
	if err != nil {
		return
	}
	shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
	log.AddHook(otellogrus.NewHook("github.com/cox96de/runner/cmd/server",
		otellogrus.WithLoggerProvider(loggerProvider)))
	trace.Init()
	return
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTraceProvider(ctx context.Context) (*sdktrace.TracerProvider, error) {
	var (
		exporter sdktrace.SpanExporter
		err      error
	)
	exporter, err = otlptracehttp.New(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create otlp trace exporter")
	}

	traceProvider := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter))
	return traceProvider, nil
}

func newMeterProvider(ctx context.Context) (*sdkmetric.MeterProvider, error) {
	metricExporter, err := otlpmetrichttp.New(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create otlp metric exporter")
	}
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)))
	return meterProvider, nil
}

func newLoggerProvider(ctx context.Context) (*otellog.LoggerProvider, error) {
	logExporter, err := otlploghttp.New(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create otlp log exporter")
	}
	loggerProvider := otellog.NewLoggerProvider(
		otellog.WithProcessor(otellog.NewBatchProcessor(logExporter)),
	)
	return loggerProvider, nil
}

package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
)

func newTraceProvider(ctx context.Context, endpoint string) (*trace.TracerProvider, error) {
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			attribute.String("service.name", "periodic-trace-generator"),
		),
	)
	if err != nil {
		return nil, err
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	)

	otel.SetTracerProvider(tp)

	return tp, nil
}

func main() {
	endpoint := flag.String("endpoint", "localhost:4317", "OTLP gRPC endpoint")
	interval := flag.Duration("interval", 5*time.Second, "Trace generation interval")

	flag.Parse()

	ctx, signalCancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer signalCancel()

	tp, err := newTraceProvider(ctx, *endpoint)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		shutdownCtx, cancel := context.WithTimeout(
			// We should not use the context returned from signal.NotifyContext.
			// If so, TraceProvidor won't be correctly shut down in case of SIGTERM.
			context.Background(),
			5*time.Second,
		)
		defer cancel()

		if err := tp.Shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown error: %v", err)
		}
	}()

	tracer := otel.Tracer("")

	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	log.Println("start periodic tracing")

loop:
	for {
		select {
		case <-ticker.C:
			func() {
				ctx, parent := tracer.Start(ctx, "parent")
				defer parent.End()
				time.Sleep(300 * time.Millisecond)

				_, child := tracer.Start(ctx, "child")
				time.Sleep(200 * time.Millisecond)
				child.End()

				time.Sleep(300 * time.Millisecond)
			}()
		case <-ctx.Done():
			break loop
		}
	}

	log.Println("stop periodic tracing")
}

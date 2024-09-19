package middlewares

import (
	"context"
	"fmt"
	"log"


	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// setupOTelSDK bootstraps the OpenTelemetry pipeline.
// SetupOTelSDK initializes OpenTelemetry with the OTLP exporter for tracing.
func SetupOTelSDK(ctx context.Context) (func(context.Context) error, error) {
	// Create a new OTLP HTTP exporter
	client := otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint("localhost:4318"), // Replace with your OTLP collector endpoint
		otlptracehttp.WithInsecure(),                // Use WithInsecure if not using TLS
	)

	exporter, err := otlptrace.New(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
	}

	// Create a resource to describe this service
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("GoWebServer"), // Customize with your service name
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Set up the tracer provider with the exporter and resource
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter), // Use the exporter to export spans
		trace.WithResource(res),     // Attach resource information to the traces
	)

	// Set the global tracer provider
	otel.SetTracerProvider(tp)

	// Function to shutdown the tracer provider
	shutdown := func(ctx context.Context) error {
		// Ensure all spans are exported before shutting down
		err := tp.Shutdown(ctx)
		if err != nil {
			log.Printf("failed to shutdown tracer provider: %v", err)
		}
		return err
	}

	return shutdown, nil
}

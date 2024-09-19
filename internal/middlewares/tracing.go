package middlewares

import (
	"context"
	"fmt"

	"github.com/deezer/groroti/internal/config"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

var TP *sdktrace.TracerProvider

// SetupOTelSDK initializes OpenTelemetry with the OTLP exporter for tracing.
func SetupOTelSDK(ctx context.Context, config config.Config) (func(context.Context) error, error) {
	log.Info().Msgf("enable OpenTelemetry: %t", config.EnableTracing)
	if config.EnableTracing {
		// Create a new OTLP HTTP exporter
		log.Info().Msgf("sending OpenTelemetry traces to: %s", config.OTLPEndpoint)
		client := otlptracehttp.NewClient(
			otlptracehttp.WithEndpoint(config.OTLPEndpoint), // Replace with your OTLP collector endpoint
			otlptracehttp.WithInsecure(),                // Use WithInsecure if not using TLS
		)

		exporter, err := otlptrace.New(ctx, client)
		if err != nil {
			return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
		}

		res, err := resource.New(ctx,
			resource.WithAttributes(
				semconv.ServiceNameKey.String("GroROTI"),
			),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create resource: %w", err)
		}

		// Set up the TracerProvider with the exporter and resource
		TP = sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(res),
		)

		// Set the global TracerProvider
		otel.SetTracerProvider(TP)

		// Function to shutdown the tracer provider
		shutdown := func(ctx context.Context) error {
			// Ensure all spans are exported before shutting down
			err := TP.Shutdown(ctx)
			if err != nil {
				log.Printf("failed to shutdown tracer provider: %v", err)
			}
			return err
		}

		return shutdown, nil
	}
	return nil, nil
}

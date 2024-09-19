package middlewares

import (
	"net/http"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// MiddlewareChain Is the standard middleware chain to apply to all routes. Change this to add a middleware for all routes
func MiddlewareChain(routeName string, next http.Handler) http.Handler {
	// Wrap the handler with OpenTelemetry tracing
	tracedHandler := otelhttp.NewHandler(next, routeName)

	// Chain other middlewares as needed
	return PrometheusInstrumentation(routeName, // Ensure Prometheus instrumentation first
		VersionHeaderResponseMiddleware(tracedHandler), // Add tracing
	)
}
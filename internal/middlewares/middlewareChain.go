package middlewares

import "net/http"

// MiddlewareChain Is the standard middleware chain to apply to all routes. Change this to add a middleware for all routes
func MiddlewareChain(routeName string, next http.Handler) http.Handler {
	return PrometheusInstrumentation(routeName, //We put this one first to ensure timing is as reliable as possible
		VersionHeaderResponseMiddleware(next),
	)
}

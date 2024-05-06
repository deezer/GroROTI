package middlewares

import (
	"net/http"
)

var Version string = "development" // default value, overridden at build time

// VersionHeaderResponseMiddleware sets an "x-version" header to the response.
func VersionHeaderResponseMiddleware(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("x-version", Version)
		next.ServeHTTP(w, r)
	})
}

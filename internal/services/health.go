package services

import (
	"net/http"

	"github.com/dimiro1/health"
)

// NewHealthHandler creates the handler for service health check
func NewHealthHandler() http.Handler {
	return health.NewHandler()
}

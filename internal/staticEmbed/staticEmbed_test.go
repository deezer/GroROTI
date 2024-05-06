package staticEmbed

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLoadTemplates(t *testing.T) {
	err := LoadTemplates()
	if err != nil {
		t.Fatalf("LoadTemplates() error: %v", err)
	}

	// Verify that the Templates map is populated
	if Templates == nil {
		t.Fatalf("Templates map is nil")
	}

	// Check if a specific template exists in the map
	if _, ok := Templates["templates/index.html"]; !ok {
		t.Fatalf("Expected template not found in Templates map")
	}
}

func TestServeStatic(t *testing.T) {
	// Create a sub-file system for embedded static files
	staticFS, err := fs.Sub(EmbeddedStatic, "static")
	if err != nil {
		t.Fatalf("Unable to load staticFS: %v", err)
	}

	staticFileServer := http.FileServer(http.FS(staticFS))
	http.Handle("/static/", http.StripPrefix("/static/", staticFileServer))

	// Create a test HTTP request
	req := httptest.NewRequest("GET", "/static/simple.css", nil)
	rec := httptest.NewRecorder()

	// Serve the request
	http.DefaultServeMux.ServeHTTP(rec, req)

	// Check the response status code
	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status OK, got %d", rec.Code)
	}

	// Check the response body
	responseBody := rec.Body.String()
	if !strings.Contains(responseBody, "/* Global variables. */") {
		t.Fatalf("Expected '/* Global variables. */' in static.css file not found")
	}
}

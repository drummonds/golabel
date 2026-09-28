package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFaviconServed(t *testing.T) {
	rec := httptest.NewRecorder()
	newMux().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/favicon.svg", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "image/svg+xml" {
		t.Errorf("Content-Type = %q, want image/svg+xml", ct)
	}
	if !strings.Contains(rec.Body.String(), "<svg") {
		t.Error("body is not an SVG")
	}
}

package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBuildApp_Routes_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		method      string
		path        string
		body        string
		expectNot   int
	}{
		{name: "POST /sabotage registered", method: http.MethodPost, path: "/sabotage", expectNot: http.StatusMethodNotAllowed},
		{name: "POST /plan/sabotage registered", method: http.MethodPost, path: "/plan/sabotage", expectNot: http.StatusMethodNotAllowed},
		{name: "GET /plan/config registered", method: http.MethodGet, path: "/plan/config", expectNot: http.StatusMethodNotAllowed},
		{name: "GET /sabotage/exec registered", method: http.MethodGet, path: "/sabotage/exec", expectNot: http.StatusMethodNotAllowed},
		{name: "GET /plan/exec registered", method: http.MethodGet, path: "/plan/exec", expectNot: http.StatusMethodNotAllowed},
		{name: "GET /sabotage/config registered", method: http.MethodGet, path: "/sabotage/config", expectNot: http.StatusMethodNotAllowed},
	}

	app := buildApp()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var req *http.Request
			if tc.body != "" {
				req = httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tc.method, tc.path, nil)
			}

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode == tc.expectNot {
				t.Fatalf("route %s %s not registered (got %d)", tc.method, tc.path, resp.StatusCode)
			}
		})
	}
}
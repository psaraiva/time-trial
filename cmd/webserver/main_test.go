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
		name      string
		method    string
		path      string
		body      string
		expectNot int
	}{
		{name: "POST /time-trial registered", method: http.MethodPost, path: "/time-trial", expectNot: http.StatusMethodNotAllowed},
		{name: "GET /time-trial/config registered", method: http.MethodGet, path: "/time-trial/config", expectNot: http.StatusMethodNotAllowed},
		{name: "POST /plan registered", method: http.MethodPost, path: "/plan", expectNot: http.StatusMethodNotAllowed},
		{name: "GET /plan/config registered", method: http.MethodGet, path: "/plan/config", expectNot: http.StatusMethodNotAllowed},
		{name: "GET /sabotage registered", method: http.MethodGet, path: "/sabotage", expectNot: http.StatusMethodNotAllowed},
		{name: "GET /plan/sabotage registered", method: http.MethodGet, path: "/plan/sabotage", expectNot: http.StatusMethodNotAllowed},
		{name: "POST /param-resp registered", method: http.MethodPost, path: "/param-resp", expectNot: http.StatusMethodNotAllowed},
		{name: "GET /param-resp/config registered", method: http.MethodGet, path: "/param-resp/config", expectNot: http.StatusMethodNotAllowed},
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

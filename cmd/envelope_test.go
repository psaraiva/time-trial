package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

func newEnvelopeApp(handler fiber.Handler) *fiber.App {
	app := fiber.New()
	app.Use(envelope())
	app.Get("/", handler)
	return app
}

func TestEnvelope_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		handler       fiber.Handler
		expectStatus  int
		validateBody  func(t *testing.T, body map[string]any)
		expectRawBody string
	}{
		{
			name: "json response is wrapped in envelope",
			handler: func(c *fiber.Ctx) error {
				return c.JSON(fiber.Map{"ok": true})
			},
			expectStatus: http.StatusOK,
			validateBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				if _, ok := body["data"]; !ok {
					t.Fatal("expected data key in envelope")
				}
				if _, ok := body["duration"]; !ok {
					t.Fatal("expected duration key in envelope")
				}
				ts, ok := body["timestamp"]
				if !ok {
					t.Fatal("expected timestamp key in envelope")
				}
				if _, err := time.Parse(time.RFC3339, ts.(string)); err != nil {
					t.Fatalf("timestamp is not RFC3339: %v", err)
				}
				data := body["data"].(map[string]any)
				if got := data["ok"].(bool); !got {
					t.Fatal("expected data.ok=true")
				}
			},
		},
		{
			name: "non-json body is returned unchanged",
			handler: func(c *fiber.Ctx) error {
				c.Response().Header.SetContentType("text/plain")
				_, err := c.Response().BodyWriter().Write([]byte("plain text"))
				return err
			},
			expectStatus:  http.StatusOK,
			expectRawBody: "plain text",
		},
		{
			name: "handler error is propagated",
			handler: func(c *fiber.Ctx) error {
				return errors.New("handler failure")
			},
			expectStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := newEnvelopeApp(tc.handler)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != tc.expectStatus {
				t.Fatalf("expected status %d, got %d", tc.expectStatus, resp.StatusCode)
			}

			if tc.expectRawBody != "" {
				buf := make([]byte, len(tc.expectRawBody))
				if _, err := resp.Body.Read(buf); err != nil && err.Error() != "EOF" {
					t.Fatalf("unexpected read error: %v", err)
				}
				if string(buf) != tc.expectRawBody {
					t.Fatalf("expected body %q, got %q", tc.expectRawBody, string(buf))
				}
				return
			}

			if tc.validateBody != nil {
				var body map[string]any
				if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
					t.Fatalf("unexpected decode error: %v", err)
				}
				tc.validateBody(t, body)
			}
		})
	}
}

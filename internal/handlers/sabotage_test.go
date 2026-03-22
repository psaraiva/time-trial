package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/psaraiva/time-trial/internal/entities"
)

func newSabotageApp(t *testing.T) (*fiber.App, *entities.State) {
	t.Helper()
	state := entities.NewState()
	h := NewSabotageHandler(state)
	app := fiber.New()
	app.Post("/sabotage", h.SetSabotage)
	return app, state
}

func TestSabotageHandlerSetSabotage_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		body            string
		expectStatus    int
		expectSabotaged bool
		expectCode      int32
		expectDelayMin  int32
		expectDelayMax  int32
		expectErrKey    bool
	}{
		{
			name:            "empty body resets state",
			body:            "",
			expectStatus:    http.StatusOK,
			expectSabotaged: false,
			expectCode:      0,
			expectDelayMin:  0,
			expectDelayMax:  0,
		},
		{
			name:            "code 0 resets state",
			body:            `{"code":0}`,
			expectStatus:    http.StatusOK,
			expectSabotaged: false,
			expectCode:      0,
			expectDelayMin:  0,
			expectDelayMax:  0,
		},
		{
			name:            "code 200 sets sabotage",
			body:            `{"code":200}`,
			expectStatus:    http.StatusOK,
			expectSabotaged: true,
			expectCode:      200,
			expectDelayMin:  0,
			expectDelayMax:  0,
		},
		{
			name:            "code 400 sets sabotage",
			body:            `{"code":400}`,
			expectStatus:    http.StatusOK,
			expectSabotaged: true,
			expectCode:      400,
			expectDelayMin:  0,
			expectDelayMax:  0,
		},
		{
			name:            "code 500 sets sabotage",
			body:            `{"code":500}`,
			expectStatus:    http.StatusOK,
			expectSabotaged: true,
			expectCode:      500,
			expectDelayMin:  0,
			expectDelayMax:  0,
		},
		{
			name:            "code 500 with valid delay",
			body:            `{"code":500,"delayMin":100,"delayMax":200}`,
			expectStatus:    http.StatusOK,
			expectSabotaged: true,
			expectCode:      500,
			expectDelayMin:  100,
			expectDelayMax:  200,
		},
		{
			name:         "invalid code returns 400",
			body:         `{"code":999}`,
			expectStatus: http.StatusBadRequest,
			expectErrKey: true,
		},
		{
			name:         "invalid delay returns 400",
			body:         `{"code":500,"delayMin":-1,"delayMax":200}`,
			expectStatus: http.StatusBadRequest,
			expectErrKey: true,
		},
		{
			name:         "malformed json returns 400",
			body:         `{invalid}`,
			expectStatus: http.StatusBadRequest,
			expectErrKey: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app, _ := newSabotageApp(t)

			var bodyReader *strings.Reader
			if tc.body != "" {
				bodyReader = strings.NewReader(tc.body)
			} else {
				bodyReader = strings.NewReader("")
			}

			req := httptest.NewRequest(http.MethodPost, "/sabotage", bodyReader)
			if tc.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != tc.expectStatus {
				t.Fatalf("expected status %d, got %d", tc.expectStatus, resp.StatusCode)
			}

			var body map[string]any
			if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
				t.Fatalf("unexpected decode error: %v", err)
			}

			if tc.expectErrKey {
				if _, ok := body["error"]; !ok {
					t.Fatal("expected error key in response body")
				}
				return
			}

			if got := body["sabotaged"].(bool); got != tc.expectSabotaged {
				t.Fatalf("expected sabotaged=%v, got %v", tc.expectSabotaged, got)
			}
			if got := int32(body["code"].(float64)); got != tc.expectCode {
				t.Fatalf("expected code=%d, got %d", tc.expectCode, got)
			}
			if got := int32(body["delayMin"].(float64)); got != tc.expectDelayMin {
				t.Fatalf("expected delayMin=%d, got %d", tc.expectDelayMin, got)
			}
			if got := int32(body["delayMax"].(float64)); got != tc.expectDelayMax {
				t.Fatalf("expected delayMax=%d, got %d", tc.expectDelayMax, got)
			}
		})
	}
}

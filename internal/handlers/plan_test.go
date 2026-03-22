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

func newPlanApp(t *testing.T) (*fiber.App, *entities.Plan) {
	t.Helper()
	plan := entities.NewPlan()
	h := NewPlanHandler(plan)
	app := fiber.New()
	app.Post("/plan", h.SetPlan)
	app.Get("/plan/config", h.GetConfig)
	return app, plan
}

func TestPlanHandlerSetPlan_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		body         string
		expectStatus int
		expectActive bool
		expectSteps  int
		expectErrKey bool
	}{
		{
			name:         "empty body clears plan",
			body:         "",
			expectStatus: http.StatusOK,
			expectActive: false,
		},
		{
			name:         "empty plan array returns 400",
			body:         `{"plan":[]}`,
			expectStatus: http.StatusBadRequest,
			expectErrKey: true,
		},
		{
			name:         "malformed json returns 400",
			body:         `{invalid}`,
			expectStatus: http.StatusBadRequest,
			expectErrKey: true,
		},
		{
			name:         "invalid code in step returns 400",
			body:         `{"plan":[{"code":999,"delayMin":0,"delayMax":0}]}`,
			expectStatus: http.StatusBadRequest,
			expectErrKey: true,
		},
		{
			name:         "invalid delay in step returns 400",
			body:         `{"plan":[{"code":500,"delayMin":-1,"delayMax":200}]}`,
			expectStatus: http.StatusBadRequest,
			expectErrKey: true,
		},
		{
			name:         "valid single step plan",
			body:         `{"plan":[{"code":500,"delayMin":0,"delayMax":0}]}`,
			expectStatus: http.StatusOK,
			expectActive: true,
			expectSteps:  1,
		},
		{
			name:         "valid two step plan",
			body:         `{"plan":[{"code":500,"delayMin":100,"delayMax":200},{"code":200,"delayMin":0,"delayMax":0}]}`,
			expectStatus: http.StatusOK,
			expectActive: true,
			expectSteps:  2,
		},
		{
			name:         "code 200 step valid",
			body:         `{"plan":[{"code":200,"delayMin":0,"delayMax":0}]}`,
			expectStatus: http.StatusOK,
			expectActive: true,
			expectSteps:  1,
		},
		{
			name:         "code 400 step valid",
			body:         `{"plan":[{"code":400,"delayMin":0,"delayMax":0}]}`,
			expectStatus: http.StatusOK,
			expectActive: true,
			expectSteps:  1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app, _ := newPlanApp(t)

			req := httptest.NewRequest(http.MethodPost, "/plan", strings.NewReader(tc.body))
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

			if got := body["active"].(bool); got != tc.expectActive {
				t.Fatalf("expected active=%v, got %v", tc.expectActive, got)
			}

			if tc.expectActive {
				if got := int(body["steps"].(float64)); got != tc.expectSteps {
					t.Fatalf("expected steps=%d, got %d", tc.expectSteps, got)
				}
			}
		})
	}
}

func TestPlanHandlerGetConfig_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		setup        func(*entities.Plan)
		expectStatus int
		expectActive bool
		expectSteps  []struct{ code, delayMin, delayMax int32 }
	}{
		{
			name:         "no active plan returns 404",
			setup:        func(p *entities.Plan) {},
			expectStatus: http.StatusNotFound,
			expectActive: false,
		},
		{
			name: "active single step plan",
			setup: func(p *entities.Plan) {
				s := entities.NewState()
				s.SetCode(500)
				p.Set([]*entities.State{s})
			},
			expectStatus: http.StatusOK,
			expectActive: true,
			expectSteps:  []struct{ code, delayMin, delayMax int32 }{{500, 0, 0}},
		},
		{
			name: "active two step plan with delay",
			setup: func(p *entities.Plan) {
				s1 := entities.NewState()
				s1.SetCode(500)
				if err := s1.SetDelay(100, 200); err != nil {
					t.Fatalf("unexpected SetDelay error: %v", err)
				}
				s2 := entities.NewState()
				s2.SetCode(200)
				p.Set([]*entities.State{s1, s2})
			},
			expectStatus: http.StatusOK,
			expectActive: true,
			expectSteps: []struct{ code, delayMin, delayMax int32 }{
				{500, 100, 200},
				{200, 0, 0},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app, plan := newPlanApp(t)
			tc.setup(plan)

			req := httptest.NewRequest(http.MethodGet, "/plan/config", nil)
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

			if !tc.expectActive {
				if _, ok := body["error"]; !ok {
					t.Fatal("expected error key in response body")
				}
				return
			}

			if got := body["active"].(bool); !got {
				t.Fatal("expected active=true")
			}

			steps := body["steps"].([]any)
			if len(steps) != len(tc.expectSteps) {
				t.Fatalf("expected %d steps, got %d", len(tc.expectSteps), len(steps))
			}

			for i, step := range steps {
				m := step.(map[string]any)
				if got := int32(m["code"].(float64)); got != tc.expectSteps[i].code {
					t.Fatalf("step %d: expected code=%d, got %d", i, tc.expectSteps[i].code, got)
				}
				if got := int32(m["delayMin"].(float64)); got != tc.expectSteps[i].delayMin {
					t.Fatalf("step %d: expected delayMin=%d, got %d", i, tc.expectSteps[i].delayMin, got)
				}
				if got := int32(m["delayMax"].(float64)); got != tc.expectSteps[i].delayMax {
					t.Fatalf("step %d: expected delayMax=%d, got %d", i, tc.expectSteps[i].delayMax, got)
				}
			}
		})
	}
}

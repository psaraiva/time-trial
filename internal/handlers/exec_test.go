package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/psaraiva/time-trial/internal/entities"
)

var validExecCodes = map[int]bool{200: true, 400: true, 500: true}

func newExecApp(t *testing.T) (*fiber.App, *entities.State, *entities.Plan, *entities.ParamResp) {
	t.Helper()
	state := entities.NewState()
	plan := entities.NewPlan()
	paramResp := entities.NewParamResp()
	h := NewExecHandler(state, plan, paramResp)
	app := fiber.New()
	app.Get("/exec", h.Exec)
	app.Get("/exec/plan", h.ExecPlan)
	return app, state, plan, paramResp
}

func TestExecHandlerExec_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		setup           func(*entities.State)
		expectCode      int
		expectRandom    bool
		expectSabotaged bool
	}{
		{
			name:            "code 0 returns random valid code",
			setup:           func(s *entities.State) {},
			expectRandom:    true,
			expectSabotaged: false,
		},
		{
			name: "code 200 returns 200",
			setup: func(s *entities.State) {
				s.SetCode(200)
			},
			expectCode:      200,
			expectSabotaged: true,
		},
		{
			name: "code 400 returns 400",
			setup: func(s *entities.State) {
				s.SetCode(400)
			},
			expectCode:      400,
			expectSabotaged: true,
		},
		{
			name: "code 500 returns 500",
			setup: func(s *entities.State) {
				s.SetCode(500)
			},
			expectCode:      500,
			expectSabotaged: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app, state, _, _ := newExecApp(t)
			tc.setup(state)

			req := httptest.NewRequest(http.MethodGet, "/exec", nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if tc.expectRandom {
				if !validExecCodes[resp.StatusCode] {
					t.Fatalf("expected one of [200,400,500], got %d", resp.StatusCode)
				}
			} else {
				if resp.StatusCode != tc.expectCode {
					t.Fatalf("expected status %d, got %d", tc.expectCode, resp.StatusCode)
				}
			}

			var body map[string]any
			if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
				t.Fatalf("unexpected decode error: %v", err)
			}

			if _, ok := body["code"]; !ok {
				t.Fatal("expected code key in response body")
			}

			sabotaged, ok := body["sabotaged"]
			if !ok {
				t.Fatal("expected sabotaged key in response body")
			}
			if sabotaged.(bool) != tc.expectSabotaged {
				t.Fatalf("expected sabotaged=%v, got %v", tc.expectSabotaged, sabotaged)
			}
		})
	}
}

func TestExecHandlerExecPlan_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		setup        func(*entities.Plan)
		expectStatus int
		expectCode   int32
		expectErrMsg string
	}{
		{
			name:         "no active plan returns 404",
			setup:        func(p *entities.Plan) {},
			expectStatus: http.StatusNotFound,
			expectErrMsg: "no active plan or all steps have been consumed",
		},
		{
			name: "cancelled plan returns 404 with interrupted message",
			setup: func(p *entities.Plan) {
				s := entities.NewState()
				s.SetCode(500)
				p.Set([]*entities.State{s})
				p.Clear()
			},
			expectStatus: http.StatusNotFound,
			expectErrMsg: "plan interrupted",
		},
		{
			name: "active plan returns next step code",
			setup: func(p *entities.Plan) {
				s := entities.NewState()
				s.SetCode(500)
				p.Set([]*entities.State{s})
			},
			expectStatus: http.StatusInternalServerError,
			expectCode:   500,
		},
		{
			name: "active plan with code 200",
			setup: func(p *entities.Plan) {
				s := entities.NewState()
				s.SetCode(200)
				p.Set([]*entities.State{s})
			},
			expectStatus: http.StatusOK,
			expectCode:   200,
		},
		{
			name: "exhausted plan returns 404",
			setup: func(p *entities.Plan) {
				s := entities.NewState()
				s.SetCode(200)
				p.Set([]*entities.State{s})
				p.Next()
			},
			expectStatus: http.StatusNotFound,
			expectErrMsg: "no active plan or all steps have been consumed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app, _, plan, _ := newExecApp(t)
			tc.setup(plan)

			req := httptest.NewRequest(http.MethodGet, "/exec/plan", nil)
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

			if tc.expectErrMsg != "" {
				errVal, ok := body["error"]
				if !ok {
					t.Fatal("expected error key in response body")
				}
				if errVal.(string) != tc.expectErrMsg {
					t.Fatalf("expected error %q, got %q", tc.expectErrMsg, errVal.(string))
				}
				return
			}

			if got := int32(body["code"].(float64)); got != tc.expectCode {
				t.Fatalf("expected code=%d, got %d", tc.expectCode, got)
			}

			sabotaged, ok := body["sabotaged"]
			if !ok {
				t.Fatal("expected sabotaged key in response body")
			}
			if sabotaged.(bool) != true {
				t.Fatalf("expected sabotaged=true, got %v", sabotaged)
			}
		})
	}
}

func TestExecHandlerExec_ParamResp(t *testing.T) {
	t.Parallel()

	paramConfig := &entities.ResponseConfig{
		StatusCode: 200,
		Item: entities.ItemConfig{
			IsCollection: false,
			Properties: []entities.Property{
				{
					PropertyBase: entities.PropertyBase{
						Name: "label", Type: entities.PropertyTypeString,
						IsRequired: true, MinLength: 3, MaxLength: 6,
					},
					PropertyString: &entities.PropertyStringConfig{Chars: "abcABC"},
				},
			},
		},
	}

	tests := []struct {
		name            string
		stateCode       int32
		paramRespActive bool
		expectStatus    int
		expectGenerated bool
	}{
		{
			name:            "code 200 with param-resp active returns generated body",
			stateCode:       200,
			paramRespActive: true,
			expectStatus:    http.StatusOK,
			expectGenerated: true,
		},
		{
			name:            "code 200 without param-resp returns standard body",
			stateCode:       200,
			paramRespActive: false,
			expectStatus:    http.StatusOK,
			expectGenerated: false,
		},
		{
			name:            "code 500 with param-resp active returns standard body",
			stateCode:       500,
			paramRespActive: true,
			expectStatus:    http.StatusInternalServerError,
			expectGenerated: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app, state, _, pr := newExecApp(t)
			state.SetCode(tc.stateCode)
			if tc.paramRespActive {
				pr.Set(paramConfig)
			}

			req := httptest.NewRequest(http.MethodGet, "/exec", nil)
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

			if tc.expectGenerated {
				if _, ok := body["label"]; !ok {
					t.Fatal("expected generated field 'label' in body")
				}
				if _, ok := body["sabotaged"]; ok {
					t.Fatal("expected no 'sabotaged' key in generated body")
				}
			} else {
				if _, ok := body["sabotaged"]; !ok {
					t.Fatal("expected 'sabotaged' key in standard body")
				}
			}
		})
	}
}

func TestExecHandlerExecPlan_ParamResp(t *testing.T) {
	t.Parallel()

	paramConfig := &entities.ResponseConfig{
		StatusCode: 200,
		Item: entities.ItemConfig{
			IsCollection: true,
			Quantity:     2,
			Properties: []entities.Property{
				{
					PropertyBase: entities.PropertyBase{
						Name: "id", Type: entities.PropertyTypeInt,
						IsRequired: true, MinLength: 1, MaxLength: 100,
					},
					PropertyInt: &entities.PropertyIntConfig{IsAcceptNegativeValue: false},
				},
			},
		},
	}

	tests := []struct {
		name            string
		planCode        int32
		paramRespActive bool
		expectStatus    int
		expectArray     bool
	}{
		{
			name:            "plan code 200 with param-resp returns generated collection",
			planCode:        200,
			paramRespActive: true,
			expectStatus:    http.StatusOK,
			expectArray:     true,
		},
		{
			name:            "plan code 200 without param-resp returns standard body",
			planCode:        200,
			paramRespActive: false,
			expectStatus:    http.StatusOK,
			expectArray:     false,
		},
		{
			name:            "plan code 400 with param-resp returns standard body",
			planCode:        400,
			paramRespActive: true,
			expectStatus:    http.StatusBadRequest,
			expectArray:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app, _, plan, pr := newExecApp(t)
			s := entities.NewState()
			s.SetCode(tc.planCode)
			plan.Set([]*entities.State{s})
			if tc.paramRespActive {
				pr.Set(paramConfig)
			}

			req := httptest.NewRequest(http.MethodGet, "/exec/plan", nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != tc.expectStatus {
				t.Fatalf("expected status %d, got %d", tc.expectStatus, resp.StatusCode)
			}

			if tc.expectArray {
				var arr []any
				if err := json.NewDecoder(resp.Body).Decode(&arr); err != nil {
					t.Fatalf("expected array body, decode error: %v", err)
				}
				if len(arr) != 2 {
					t.Fatalf("expected 2 items in collection, got %d", len(arr))
				}
			} else {
				var body map[string]any
				if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
					t.Fatalf("unexpected decode error: %v", err)
				}
				if _, ok := body["error"]; ok {
					return
				}
				if _, ok := body["sabotaged"]; !ok {
					t.Fatal("expected 'sabotaged' key in standard body")
				}
			}
		})
	}
}

func TestApplyDelay_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		min  int32
		max  int32
	}{
		{name: "no delay", min: 0, max: 0},
		{name: "fixed delay min equals max", min: 1, max: 1},
		{name: "range delay min less than max", min: 1, max: 2},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := entities.NewState()
			if err := s.SetDelay(tc.min, tc.max); err != nil {
				t.Fatalf("unexpected SetDelay error: %v", err)
			}

			// applyDelay must not panic or block indefinitely
			applyDelay(s)
		})
	}
}

func TestExecHandlerExecWithDelay_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		code       int32
		min        int32
		max        int32
		expectCode int
	}{
		{
			name:       "exec with fixed delay",
			code:       200,
			min:        1,
			max:        1,
			expectCode: http.StatusOK,
		},
		{
			name:       "exec plan with delay range",
			code:       500,
			min:        1,
			max:        2,
			expectCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app, state, _, _ := newExecApp(t)
			state.SetCode(tc.code)
			if err := state.SetDelay(tc.min, tc.max); err != nil {
				t.Fatalf("unexpected SetDelay error: %v", err)
			}

			req := httptest.NewRequest(http.MethodGet, "/exec", nil)
			resp, err := app.Test(req, 5000)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != tc.expectCode {
				t.Fatalf("expected status %d, got %d", tc.expectCode, resp.StatusCode)
			}
		})
	}
}

func TestExecPlanWithDelay_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		code       int32
		min        int32
		max        int32
		expectCode int
	}{
		{
			name:       "exec plan step with fixed delay",
			code:       200,
			min:        1,
			max:        1,
			expectCode: http.StatusOK,
		},
		{
			name:       "exec plan step with delay range",
			code:       400,
			min:        1,
			max:        2,
			expectCode: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app, _, plan, _ := newExecApp(t)
			s := entities.NewState()
			s.SetCode(tc.code)
			if err := s.SetDelay(tc.min, tc.max); err != nil {
				t.Fatalf("unexpected SetDelay error: %v", err)
			}
			plan.Set([]*entities.State{s})

			req := httptest.NewRequest(http.MethodGet, "/exec/plan", nil)
			resp, err := app.Test(req, 5000)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != tc.expectCode {
				t.Fatalf("expected status %d, got %d", tc.expectCode, resp.StatusCode)
			}
		})
	}
}

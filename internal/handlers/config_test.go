package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/psaraiva/time-trial/internal/entities"
)

func TestConfigHandlerGetConfig_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		setup           func(*entities.State)
		expectSabotaged bool
		expectCode      int32
		expectDelayMin  int32
		expectDelayMax  int32
	}{
		{
			name:            "no sabotage",
			setup:           func(s *entities.State) {},
			expectSabotaged: false,
			expectCode:      0,
			expectDelayMin:  0,
			expectDelayMax:  0,
		},
		{
			name: "code 500 sabotaged",
			setup: func(s *entities.State) {
				s.SetCode(500)
			},
			expectSabotaged: true,
			expectCode:      500,
			expectDelayMin:  0,
			expectDelayMax:  0,
		},
		{
			name: "code 200 with delay",
			setup: func(s *entities.State) {
				s.SetCode(200)
				if err := s.SetDelay(100, 200); err != nil {
					t.Fatalf("unexpected SetDelay error: %v", err)
				}
			},
			expectSabotaged: true,
			expectCode:      200,
			expectDelayMin:  100,
			expectDelayMax:  200,
		},
		{
			name: "code 400",
			setup: func(s *entities.State) {
				s.SetCode(400)
			},
			expectSabotaged: true,
			expectCode:      400,
			expectDelayMin:  0,
			expectDelayMax:  0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			state := entities.NewState()
			tc.setup(state)

			h := NewConfigHandler(state)
			app := fiber.New()
			app.Get("/config", h.GetConfig)

			req := httptest.NewRequest(http.MethodGet, "/config", nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected status 200, got %d", resp.StatusCode)
			}

			var body map[string]any
			if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
				t.Fatalf("unexpected decode error: %v", err)
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

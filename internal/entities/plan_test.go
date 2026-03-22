package entities

import "testing"

func makeState(t *testing.T, code, min, max int32) *State {
	t.Helper()

	s := NewState()
	s.SetCode(code)
	if err := s.SetDelay(min, max); err != nil {
		t.Fatalf("unexpected SetDelay error: %v", err)
	}
	return s
}

func TestNewPlan_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "new plan starts empty and not cancelled"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := NewPlan()
			if p == nil {
				t.Fatal("expected non-nil plan")
			}
			if p.IsCancelled() {
				t.Fatal("expected new plan to be not cancelled")
			}
			if states, ok := p.IsActive(); ok || states != nil {
				t.Fatalf("expected no active plan, got ok=%v states=%v", ok, states)
			}
		})
	}
}

func TestPlanSetAndIsActive_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		states         []*State
		expectedActive bool
		expectedLen    int
		cancelBefore   bool
	}{
		{name: "set nil plan", states: nil, expectedActive: false, expectedLen: 0, cancelBefore: false},
		{name: "set empty plan", states: []*State{}, expectedActive: false, expectedLen: 0, cancelBefore: false},
		{
			name:           "set two states clears cancelled and activates",
			states:         []*State{makeState(t, 500, 100, 200), makeState(t, 400, 300, 400)},
			expectedActive: true,
			expectedLen:    2,
			cancelBefore:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := NewPlan()
			if tc.cancelBefore {
				p.Clear()
				if !p.IsCancelled() {
					t.Fatal("expected plan cancelled after Clear")
				}
			}

			p.Set(tc.states)

			if p.IsCancelled() {
				t.Fatal("expected Set to clear cancelled flag")
			}

			states, ok := p.IsActive()
			if ok != tc.expectedActive {
				t.Fatalf("expected active=%v, got %v", tc.expectedActive, ok)
			}
			if len(states) != tc.expectedLen {
				t.Fatalf("expected len=%d, got %d", tc.expectedLen, len(states))
			}
			if tc.expectedActive {
				states[0] = nil
				nextSnapshot, nextOK := p.IsActive()
				if !nextOK {
					t.Fatal("expected active plan on second snapshot")
				}
				if nextSnapshot[0] == nil {
					t.Fatal("expected IsActive to return a copied slice")
				}
			}
		})
	}
}

func TestPlanNext_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setup       func(*Plan)
		expectState *State
		expectOK    bool
	}{
		{
			name:        "next on empty plan",
			setup:       func(p *Plan) {},
			expectState: nil,
			expectOK:    false,
		},
		{
			name: "next on cancelled plan",
			setup: func(p *Plan) {
				p.Clear()
			},
			expectState: nil,
			expectOK:    false,
		},
		{
			name: "next returns first state",
			setup: func(p *Plan) {
				s1 := makeState(t, 500, 100, 200)
				s2 := makeState(t, 400, 300, 400)
				p.Set([]*State{s1, s2})
			},
			expectState: makeState(t, 500, 100, 200),
			expectOK:    true,
		},
		{
			name: "next after exhausting plan",
			setup: func(p *Plan) {
				s1 := makeState(t, 200, 100, 100)
				p.Set([]*State{s1})
				_, _ = p.Next()
			},
			expectState: nil,
			expectOK:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := NewPlan()
			tc.setup(p)

			got, ok := p.Next()
			if ok != tc.expectOK {
				t.Fatalf("expected ok=%v, got %v", tc.expectOK, ok)
			}
			if !tc.expectOK {
				if got != nil {
					t.Fatalf("expected nil state, got %v", got)
				}
				return
			}

			if got == nil {
				t.Fatal("expected non-nil state")
			}

			if got.GetCode() != tc.expectState.GetCode() {
				t.Fatalf("expected code=%d, got %d", tc.expectState.GetCode(), got.GetCode())
			}
			gotMin, gotMax := got.GetDelay()
			expectedMin, expectedMax := tc.expectState.GetDelay()
			if gotMin != expectedMin || gotMax != expectedMax {
				t.Fatalf("expected delay (%d, %d), got (%d, %d)", expectedMin, expectedMax, gotMin, gotMax)
			}
		})
	}
}

func TestPlanClear_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "clear interrupts active plan"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := NewPlan()
			s1 := makeState(t, 500, 100, 200)
			p.Set([]*State{s1})

			p.Clear()

			if !p.IsCancelled() {
				t.Fatal("expected cancelled=true after Clear")
			}
			if states, ok := p.IsActive(); ok || states != nil {
				t.Fatalf("expected no active plan after Clear, got ok=%v states=%v", ok, states)
			}
			if s, ok := p.Next(); ok || s != nil {
				t.Fatalf("expected Next to be interrupted after Clear, got ok=%v state=%v", ok, s)
			}
		})
	}
}

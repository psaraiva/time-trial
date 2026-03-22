package entities

import "testing"

func TestValidateDelay_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		min    int32
		max    int32
		errMsg string
	}{
		{name: "no delay", min: 0, max: 0, errMsg: ""},
		{name: "valid full range", min: DelayLowerBoundMs, max: DelayUpperBoundMs, errMsg: ""},
		{name: "valid fixed delay period zero", min: 500, max: 500, errMsg: ""},
		{name: "valid min zero with max", min: 0, max: 500, errMsg: ""},
		{name: "invalid min below lower bound", min: -1, max: 10, errMsg: "delayMin must be between 0 and 60000ms"},
		{name: "invalid min above upper bound", min: DelayUpperBoundMs + 1, max: DelayUpperBoundMs + 1, errMsg: "delayMin must be between 0 and 60000ms"},
		{name: "invalid max below lower bound", min: 10, max: -1, errMsg: "delayMax must be between 0 and 60000ms"},
		{name: "invalid max above upper bound", min: 10, max: DelayUpperBoundMs + 1, errMsg: "delayMax must be between 0 and 60000ms"},
		{name: "invalid period min greater than max", min: 1000, max: 500, errMsg: "delay period must be between 0 and 60000ms"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateDelay(tc.min, tc.max)
			if tc.errMsg == "" && err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
			if tc.errMsg != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tc.errMsg)
				}
				if err.Error() != tc.errMsg {
					t.Fatalf("expected error %q, got %q", tc.errMsg, err.Error())
				}
			}
		})
	}
}

func TestStateCode_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		code int32
	}{
		{name: "code zero", code: 0},
		{name: "code 200", code: 200},
		{name: "code 500", code: 500},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := NewState()
			if s == nil {
				t.Fatal("expected non-nil state")
			}
			s.SetCode(tc.code)
			if got := s.GetCode(); got != tc.code {
				t.Fatalf("expected code %d, got %d", tc.code, got)
			}
		})
	}
}

func TestStateSetDelay_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		min         int32
		max         int32
		expectErr   bool
		expectedMin int32
		expectedMax int32
	}{
		{name: "set valid delay", min: 100, max: 200, expectErr: false, expectedMin: 100, expectedMax: 200},
		{name: "set no delay", min: 0, max: 0, expectErr: false, expectedMin: 0, expectedMax: 0},
		{name: "invalid delay keeps previous values", min: -1, max: 10, expectErr: true, expectedMin: 300, expectedMax: 400},
		{name: "invalid period keeps previous values", min: 1000, max: 500, expectErr: true, expectedMin: 300, expectedMax: 400},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := NewState()
			if tc.expectErr {
				if err := s.SetDelay(300, 400); err != nil {
					t.Fatalf("unexpected setup error: %v", err)
				}
			}

			err := s.SetDelay(tc.min, tc.max)
			if tc.expectErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tc.expectErr && err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}

			gotMin, gotMax := s.GetDelay()
			if gotMin != tc.expectedMin || gotMax != tc.expectedMax {
				t.Fatalf("expected delay (%d, %d), got (%d, %d)", tc.expectedMin, tc.expectedMax, gotMin, gotMax)
			}
		})
	}
}

func TestStateReset_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		code int32
		min  int32
		max  int32
	}{
		{name: "reset from configured state", code: 500, min: 100, max: 200},
		{name: "reset from no-delay state", code: 200, min: 0, max: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := NewState()
			s.SetCode(tc.code)
			if err := s.SetDelay(tc.min, tc.max); err != nil {
				t.Fatalf("unexpected SetDelay error: %v", err)
			}

			s.Reset()

			if got := s.GetCode(); got != 0 {
				t.Fatalf("expected code 0 after reset, got %d", got)
			}
			gotMin, gotMax := s.GetDelay()
			if gotMin != 0 || gotMax != 0 {
				t.Fatalf("expected delay (0, 0) after reset, got (%d, %d)", gotMin, gotMax)
			}
		})
	}
}

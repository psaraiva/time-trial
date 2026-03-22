package entities

import (
	"fmt"
	"sync/atomic"
)

const (
	DelayLowerBoundMs int32 = 0
	DelayUpperBoundMs int32 = 60000
)

// State holds the forced HTTP status code and optional delay range for sabotage mode.
// A code of 0 means random behavior (no sabotage).
// DelayMin and DelayMax of 0 means no delay.
type State struct {
	code     atomic.Int32
	delayMin atomic.Int32
	delayMax atomic.Int32
}

func NewState() *State {
	return &State{}
}

func (s *State) SetCode(code int32) {
	s.code.Store(code)
}

func (s *State) GetCode() int32 {
	return s.code.Load()
}

func ValidateDelay(min, max int32) error {
	if min == 0 && max == 0 {
		return nil
	}
	if min < DelayLowerBoundMs || min > DelayUpperBoundMs {
		return fmt.Errorf("delayMin must be between 0 and 60000ms")
	}
	if max < DelayLowerBoundMs || max > DelayUpperBoundMs {
		return fmt.Errorf("delayMax must be between 0 and 60000ms")
	}
	period := max - min
	if period < DelayLowerBoundMs || period > DelayUpperBoundMs {
		return fmt.Errorf("delay period must be between 0 and 60000ms")
	}
	return nil
}

func (s *State) SetDelay(min, max int32) error {
	if err := ValidateDelay(min, max); err != nil {
		return err
	}
	s.delayMin.Store(min)
	s.delayMax.Store(max)
	return nil
}

func (s *State) GetDelay() (min, max int32) {
	return s.delayMin.Load(), s.delayMax.Load()
}

func (s *State) Reset() {
	s.code.Store(0)
	s.delayMin.Store(0)
	s.delayMax.Store(0)
}

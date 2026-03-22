package entities

import (
	"sync"
	"sync/atomic"
)

// Plan holds an ordered list of sabotage States to be consumed one per request.
// Each State in the list is a full sabotage configuration (code + delay range).
// States are consumed in order; the plan remains in memory until explicitly cleared.
type Plan struct {
	mu        sync.Mutex
	states    []*State
	index     int
	cancelled atomic.Bool
}

func NewPlan() *Plan {
	return &Plan{}
}

// Set replaces the current plan, resets the cursor and clears the cancelled flag.
func (p *Plan) Set(states []*State) {
	p.cancelled.Store(false)
	p.mu.Lock()
	defer p.mu.Unlock()
	p.states = states
	p.index = 0
}

// Next returns the next State in the plan and advances the cursor.
// Returns nil, false when the plan is empty, cancelled, or all states have been consumed.
func (p *Plan) Next() (*State, bool) {
	if p.cancelled.Load() {
		return nil, false
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.states) == 0 || p.index >= len(p.states) {
		return nil, false
	}
	s := p.states[p.index]
	p.index++
	return s, true
}

// Cancelled returns true if the plan was interrupted via Clear before being exhausted.
func (p *Plan) IsCancelled() bool {
	return p.cancelled.Load()
}

// IsActive returns a copy of the full loaded plan and true if one exists,
// regardless of how many states have already been consumed.
// Returns nil, false if no plan is loaded.
func (p *Plan) IsActive() ([]*State, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.states) == 0 {
		return nil, false
	}
	snapshot := make([]*State, len(p.states))
	copy(snapshot, p.states)
	return snapshot, true
}

// Clear removes the active plan from memory and raises the cancelled flag
// so the next call to Next() is safely interrupted.
func (p *Plan) Clear() {
	p.cancelled.Store(true)
	p.mu.Lock()
	defer p.mu.Unlock()
	p.states = nil
	p.index = 0
}

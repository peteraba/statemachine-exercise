// This package provides a simple StateMachine implementation
// with two Rule types, both implementing the TransitionRule interface:
// - SimpleTransitionRule: always allows the transition between two states as long as they exist
// - ConditionalTransitionRule: allows the transition between two states only if some conditions are met
package main

import (
	"fmt"
)

var (
	TransitionNotAllowed = fmt.Errorf("error: transition not allowed")
	StateNotFound        = fmt.Errorf("error: state not found")
)

// State describes a possible state in a StateMachine
type State string

// TransitionRule allows or denies transitioning between two states
type TransitionRule interface {
	From() State
	To() State
	Valid(fromState, toState State, params ...interface{}) bool
}

// SimpleTransitionRule always allows the transition between two states as long as they exist
type SimpleTransitionRule struct {
	from State
	to   State
}

// NewSimpleTransitionRule creates a new SimpleTransitionRule
func NewSimpleTransitionRule(from, to State) *SimpleTransitionRule {
	return &SimpleTransitionRule{
		from: from,
		to:   to,
	}
}

// From retrieves the start state the transition rule applies to
func (r *SimpleTransitionRule) From() State {
	return r.from
}

// To retrieves the end state the transition rule applies to
func (r *SimpleTransitionRule) To() State {
	return r.to
}

// Valid is true if transitioning between two states is allowed
func (r *SimpleTransitionRule) Valid(from, to State, params ...interface{}) bool {
	return from == r.from && to == r.to
}

// ConditionalTransitionRule allows the transition between two states only if some conditions are met
type ConditionalTransitionRule struct {
	from      State
	to        State
	condition func(params ...interface{}) bool
}

// NewConditionalTransitionRule creates a new ConditionalTransitionRule
func NewConditionalTransitionRule(from, to State, condition func(params ...interface{}) bool) *ConditionalTransitionRule {
	return &ConditionalTransitionRule{
		from:      from,
		to:        to,
		condition: condition,
	}
}

// From retrieves the start state the transition rule applies to
func (r *ConditionalTransitionRule) From() State {
	return r.from
}

// To retrieves the end state the transition rule applies to
func (r *ConditionalTransitionRule) To() State {
	return r.to
}

// Valid is true if transitioning between two states is allowed
func (r *ConditionalTransitionRule) Valid(from, to State, params ...interface{}) bool {
	return from == r.from && to == r.to && r.condition(params...)
}

// StateMachine defines as StateMachine with current and existing states and rules to transition between states
type StateMachine struct {
	state  State
	states map[State]State
	rules  []TransitionRule
	final  bool
}

// NewStateMachine creates a new StateMachine instance
func NewStateMachine(initialState State, states ...State) *StateMachine {
	stateMap := map[State]State{
		initialState: initialState,
	}
	for _, state := range states {
		stateMap[state] = state
	}

	return &StateMachine{
		state:  initialState,
		states: stateMap,
		rules:  []TransitionRule{},
	}
}

func (sm *StateMachine) AddRule(rule TransitionRule) error {
	if sm.final {
		return fmt.Errorf("rules must be defined before finalization")
	}

	_, ok := sm.states[rule.From()]
	if !ok {
		return fmt.Errorf("state: %v, %w", rule.From(), StateNotFound)
	}

	_, ok = sm.states[rule.To()]
	if !ok {
		return fmt.Errorf("state: %v, %w", rule.To(), StateNotFound)
	}

	sm.rules = append(sm.rules, rule)

	return nil
}

// IsFinal is true if the StateMachine is ready to handle transitions
func (sm *StateMachine) IsFinal() bool {
	return sm.final
}

// State returns the current state of the StateMachine
func (sm *StateMachine) State() State {
	return sm.state
}

// Transition attempts to transition the StateMachine into a new State
// The transition is only allowed if there's a rule which allows it
func (sm *StateMachine) Transition(to State, params ...interface{}) error {
	sm.final = true

	if sm.state == to {
		return nil
	}

	_, ok := sm.states[to]
	if !ok {
		return fmt.Errorf("state: %v, %w", to, StateNotFound)
	}

	for _, rule := range sm.rules {
		if rule.From() == sm.state && rule.To() == to {
			if rule.Valid(sm.state, to, params...) {
				sm.state = to

				return nil
			}

			return TransitionNotAllowed
		}
	}

	return TransitionNotAllowed
}

// equalIntegers is a helper function to demonstrate the capabilities of the ConditionalTransitionRule
func equalIntegers(params ...interface{}) bool {
	if len(params) != 2 {
		return false
	}

	a, ok := params[0].(int)
	if !ok {
		return false
	}

	b, ok := params[1].(int)
	if !ok {
		return false
	}

	return a == b
}

// main is used for testing the StateMachine
// Initializes the StateMachine in "Initial" state
// attempts to transition into the "Canceled" state
// then transitions into the "Backlog" state
// then makes various attempts to transition into "Progress" state
// Note that the Canceled state is not added to the allowed states
// Initial -> Backlog is unconditional (SimpleTransitionRule)
// Backlog -> Progress is conditional (ConditionalTransitionRule)
func main() {
	// Initialise
	i := State("Initial")
	b := State("Backlog")
	p := State("Progress")
	c := State("Canceled")
	sm := NewStateMachine(i, b, p)
	fmt.Println("[add rule]", sm.AddRule(NewSimpleTransitionRule(i, b)))
	fmt.Println("[add rule]", sm.AddRule(NewConditionalTransitionRule(b, p, equalIntegers)))
	fmt.Println("[state]", sm.State())

	// Transition to non-existent state (Initial -> Canceled)
	fmt.Println("[transition]", sm.Transition(c))
	fmt.Println("[state]", sm.State())

	// Transition without passing rule (Initial -> Progress)
	fmt.Println("[transition]", sm.Transition(p))
	fmt.Println("[state]", sm.State())

	// Transition with passing simple rule (Initial -> Backlog)
	fmt.Println("[transition]", sm.Transition(b))
	fmt.Println("[state]", sm.State())

	// Transition with non-passing complex rule (Backlog -> Progress)
	fmt.Println("[transition]", sm.Transition(p))
	fmt.Println("[state]", sm.State())

	// Transition with non-passing complex rule II. (Backlog -> Progress)
	fmt.Println("[transition]", sm.Transition(p, 10, 15))
	fmt.Println("[state]", sm.State())

	// Transition with non-passing complex rule III. (Backlog -> Progress)
	fmt.Println("[transition]", sm.Transition(p, 10.0, 10))
	fmt.Println("[state]", sm.State())

	// Transition with passing complex rule (Backlog -> Progress)
	fmt.Println("[transition]", sm.Transition(p, 10, 10))
	fmt.Println("[state]", sm.State())
}

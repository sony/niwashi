// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package workflow

import "fmt"

type StateMachine struct {
	initial, current, final int
	states                  map[int]State
	transitions             map[int]int // from state -> to state
}

func NewStateMachine(init, final int) *StateMachine {
	if init == final {
		panic("initial state and final state cannot be the same")
	}

	sm := &StateMachine{
		initial:     init,
		current:     init,
		final:       final,
		states:      make(map[int]State),
		transitions: make(map[int]int),
	}

	s := &EmptyState{}
	sm.AddState(init, s)
	sm.AddState(final, s)

	return sm
}

func (sm *StateMachine) IsComplete() bool {
	return sm.current == sm.final
}

func (sm *StateMachine) Validate() error {
	for from, to := range sm.transitions {
		if _, exist := sm.states[from]; !exist {
			return fmt.Errorf("invalid state machine: from state %d not found", from)
		}
		if _, exist := sm.states[to]; !exist {
			return fmt.Errorf("invalid state machine: to state %d not found", to)
		}
	}

	return nil
}

func (sm *StateMachine) Execute() error {

	if sm.current == sm.initial {
		err := sm.Validate()
		if err != nil {
			return err
		}
	}

	state, exist := sm.states[sm.current]
	if !exist {
		// error
		return fmt.Errorf("no state found for current state id: %d", sm.current)
	}

	for !sm.IsComplete() {

		state.OnExecute()
		if !state.Satisfy() {
			break
		}

		// transition to next state
		state.OnExit()
		if next, exist := sm.transitions[sm.current]; exist {
			sm.current = next

			state = sm.states[sm.current]
			state.OnEnter()
		} else {
			panic(fmt.Sprintf("no transition found for current state id: %d", sm.current))
		}
	}
	return nil
}

type State interface {
	OnEnter()
	Satisfy() bool
	OnExecute()
	OnExit()
}

type EmptyState struct{}

func (s *EmptyState) OnEnter() {
}

func (s *EmptyState) Satisfy() bool {
	return true
}

func (s *EmptyState) OnExecute() {
}

func (s *EmptyState) OnExit() {
}

type defaultState struct {
	fnOnEnter   func()
	fnSatisfy   func() bool
	fnOnExecute func()
	fnOnExit    func()
}

func NewDefaultState(
	onEnter func(),
	satisfy func() bool,
	onExecute func(),
	onExit func()) State {
	return &defaultState{
		fnOnEnter:   onEnter,
		fnSatisfy:   satisfy,
		fnOnExecute: onExecute,
		fnOnExit:    onExit,
	}
}

func (s *defaultState) OnEnter() {
	if s.fnOnEnter != nil {
		s.fnOnEnter()
	}
}

func (s *defaultState) Satisfy() bool {
	if s.fnSatisfy != nil {
		return s.fnSatisfy()
	}
	return true
}

func (s *defaultState) OnExecute() {
	if s.fnOnExecute != nil {
		s.fnOnExecute()
	}
}

func (s *defaultState) OnExit() {
	if s.fnOnExit != nil {
		s.fnOnExit()
	}
}

func (sm *StateMachine) AddState(id int, state State) {
	sm.states[id] = state
}

func (sm *StateMachine) AddTransition(from, to int) {
	sm.transitions[from] = to
}

func (sm *StateMachine) AddTransitions(fromTo map[int]int) {
	for from, to := range fromTo {
		sm.transitions[from] = to
	}
}

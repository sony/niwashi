// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package workflow

type TaskAllocator interface {
	Empty() bool
	Enqueue(t Task)
	Allocate() Task
	Deallocate(t Task)
}

type taskAllocator struct {
	pendings []Task
	runnings map[string]Task
}

func NewTaskAllocator() *taskAllocator {
	return &taskAllocator{
		runnings: make(map[string]Task),
	}
}

func (ta *taskAllocator) Empty() bool {
	return len(ta.pendings) == 0 && len(ta.runnings) == 0
}

func (ta *taskAllocator) Enqueue(t Task) {
	ta.pendings = append(ta.pendings, t)
}

func (ta *taskAllocator) Allocate() Task {

	if len(ta.pendings) == 0 {
		// nothing to allocate
		return nil
	}

	// make a set of running spaces
	spaces := make(map[string]struct{})
	for k := range ta.runnings {
		spaces[k] = struct{}{}
	}

	for i := 0; i < len(ta.pendings); i++ {
		t := ta.pendings[i]

		subject := t.GetSubject()
		subSubjects := t.GetAffectSubjects()

		if allocatable(spaces, subject, subSubjects) {
			// fill runnings with subject and subSubjects(dummy)
			ta.runnings[subject] = t
			for _, s := range subSubjects {
				ta.runnings[s] = nil
			}
			// remove from pendings
			ta.pendings = append(ta.pendings[:i], ta.pendings[i+1:]...)
			return t
		}

		// fill spaces with subject and subSubjects for next iteration
		// prevent to overtake prior tasks
		spaces[subject] = struct{}{}
		for _, s := range subSubjects {
			spaces[s] = struct{}{}
		}
	}

	return nil
}

func allocatable(spaces map[string]struct{}, subject string, subSubjects []string) bool {
	if _, ok := spaces[subject]; ok {
		return false
	}

	for _, s := range subSubjects {
		if _, ok := spaces[s]; ok {
			return false
		}
	}

	return true
}
func (ta *taskAllocator) Deallocate(t Task) {
	subject := t.GetSubject()
	subSubjects := t.GetAffectSubjects()

	delete(ta.runnings, subject)
	for _, s := range subSubjects {
		delete(ta.runnings, s)
	}
}

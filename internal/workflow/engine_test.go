// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package workflow

import (
	"context"
	"log"
	"testing"
	"time"
)

type mockTask struct {
	Error     error
	sleepTime time.Duration
}

func (m *mockTask) GetName() string {
	return "mockTask"
}

func (t *mockTask) GetSubject() string {
	return "<test>"
}
func (t *mockTask) GetAffectSubjects() []string {
	return []string{}
}

func (m *mockTask) Run(_ context.Context) error {
	if m.sleepTime > 0 {
		time.Sleep(m.sleepTime)
	}
	return m.Error
}
func (m *mockTask) AfterRun(runErr error) error {
	return nil
}
func (m *mockTask) GetJob() *Job {
	return nil
}

type EmptyScheduler struct {
	DefaultErrorHolder
}

func (*EmptyScheduler) NextTask() Task {
	return nil
}
func (*EmptyScheduler) OnTaskComplete(result TaskResult) {
	panic("never called")
}
func (*EmptyScheduler) IsEmpty() bool {
	return true
}

type InfinitScheduler struct {
	DefaultErrorHolder
	sleepTime time.Duration
	count     uint
}

func (s *InfinitScheduler) NextTask() Task {
	s.count++
	return &mockTask{
		sleepTime: s.sleepTime,
	}
}

func (s *InfinitScheduler) OnTaskComplete(result TaskResult) {
	s.count--
}

func (*InfinitScheduler) IsEmpty() bool {
	return false
}

type DeadlockScheduler struct {
	DefaultErrorHolder
}

func (s *DeadlockScheduler) NextTask() Task {
	return nil
}

func (*DeadlockScheduler) OnTaskComplete(result TaskResult) {}

func (*DeadlockScheduler) IsEmpty() bool {
	return false
}

type SimpleScheduler struct {
	DefaultErrorHolder
	sleepTime time.Duration
	max       int
	count     int
	working   int
}

func newSimpleScheduler(max int, sleepTime time.Duration) *SimpleScheduler {
	return &SimpleScheduler{
		sleepTime: sleepTime,
		max:       max,
	}
}

func (s *SimpleScheduler) NextTask() Task {
	if s.count >= s.max {
		return nil
	}
	s.count++
	s.working++
	return &mockTask{
		sleepTime: s.sleepTime,
	}
}

func (s *SimpleScheduler) OnTaskComplete(result TaskResult) {
	log.Printf("task completed")
	s.working--
}

func (s *SimpleScheduler) IsEmpty() bool {
	return s.count >= s.max && s.working == 0
}

func TestEngine_Run(t *testing.T) {

	t.Run("error: no scheduler", func(t *testing.T) {
		e := NewEngine()

		err := e.Run(context.Background())
		if err == nil {
			t.Errorf("Engine.Run() error = %v, wantSomeErr ", err)
		}
	})

	t.Run("error: invalid concurrency(under min)", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Engine.Run() did not panic with invalid concurrency")
			}
		}()

		_ = NewEngine(WithMaxConcurrency(MinConcurrency - 1))
	})

	t.Run("error: invalid concurrency(over max)", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Engine.Run() did not panic with invalid concurrency")
			}
		}()

		_ = NewEngine(WithMaxConcurrency(MaxConcurrency + 1))
	})

	t.Run("error: cancel", func(t *testing.T) {
		sched := &InfinitScheduler{
			sleepTime: 800 * time.Millisecond,
		}
		concurrency := uint(3)
		e := NewEngine(
			WithScheduler(sched),
			WithMaxConcurrency(concurrency),
		)

		ctx, cancel := context.WithCancel(context.Background())
		timer := time.NewTimer(1 * time.Second)

		go func() {
			<-timer.C
			cancel()
		}()

		if sched.count != 0 {
			t.Errorf("InfinitScheduler count = %d, want 0", sched.count)
		}

		err := e.Run(ctx)
		if err == nil {
			t.Errorf("Engine.Run() error = %v, wantSomeErr", err)
		}

		if sched.count != concurrency {
			t.Errorf("InfinitScheduler count = %d, want %d", sched.count, concurrency)
		}
	})

	t.Run("error: deadlock", func(t *testing.T) {
		e := NewEngine(WithScheduler(&DeadlockScheduler{}))

		err := e.Run(context.Background())
		if err == nil {
			t.Errorf("Engine.Run() error = %v, want SomeErr", err)
		}
	})

	t.Run("success: no job", func(t *testing.T) {
		e := NewEngine(WithScheduler(&EmptyScheduler{}))

		err := e.Run(context.Background())
		if err != nil {
			t.Errorf("Engine.Run() error = %v, expect nil", err)
		}
	})

	t.Run("success: normal", func(t *testing.T) {
		tm := time.Duration(500 * time.Millisecond)

		e := NewEngine(
			WithMaxConcurrency(3),
			WithScheduler(newSimpleScheduler(3, tm)),
		)

		start := time.Now()
		err := e.Run(context.Background())
		duration := time.Since(start)

		if err != nil {
			t.Errorf("Engine.Run() error = %v, expect nil", err)
		}

		if duration < tm || duration > tm*2 {
			t.Errorf("Engine.Run() duration = %v, want around %v", duration, tm)
		}
	})

}

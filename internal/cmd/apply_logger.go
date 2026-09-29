// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/sony/niwashi/internal/logger"
	"golang.org/x/term"
)

const (
	stateSuspend int = iota
	stateRunning
)

func BuildApplyLogger(progressFormat string) (slog.Handler, *ApplyTui, error) {

	if progressFormat == "auto" {
		if term.IsTerminal(int(os.Stdout.Fd())) {
			progressFormat = "tui"
		} else {
			progressFormat = "plain"
		}
	}

	var handler slog.Handler
	options := logger.GetHandlerOptions()
	var tui *ApplyTui

	switch progressFormat {
	case "tui":
		tui = NewTui()
		handler = tui.GetHandler()
		tui.Run(context.Background())
	case "plain":
		handler = slog.NewTextHandler(os.Stdout, options)
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, options)
	case "pretty":
		handler = NewPrettyHandler()
	default:
		return nil, nil, fmt.Errorf("invalid progress format: %s", progressFormat)
	}

	return handler, tui, nil
}

type ApplyTui struct {
	Model       *Model
	View        *TuiView
	Bridge      *TuiBridgeHandler
	state       int
	done        chan struct{}
	stateChange chan int
}

func NewTui() *ApplyTui {
	m := NewModel()
	return &ApplyTui{
		Model:       m,
		View:        NewTuiView(m),
		Bridge:      NewTuiBridgeHandler(m),
		state:       stateRunning,
		done:        make(chan struct{}),
		stateChange: make(chan int),
	}
}

const refreshInterval = 200 * time.Millisecond

func (t *ApplyTui) GetHandler() slog.Handler {
	return t.Bridge
}

func (t *ApplyTui) Run(ctx context.Context) {
	updated := make(chan bool, 10)

	t.Model.OnUpdate = func() {
		updated <- true
	}

	go func() {
		ticker := time.NewTicker(refreshInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if t.state == stateRunning {
					t.View.Render()
				}
			case <-updated:
				if t.state == stateRunning {
					t.View.Render()
				}
			case <-t.done:
				if t.state == stateRunning {
					t.View.Render()
				}
				return
			case req := <-t.stateChange:
				t.state = req
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (t *ApplyTui) Pause() {
	if t == nil {
		return
	}

	t.stateChange <- stateSuspend
}

func (t *ApplyTui) Resume() {
	if t == nil {
		return
	}

	t.stateChange <- stateRunning
}

func (t *ApplyTui) Stop() {
	if t == nil {
		return
	}

	t.done <- struct{}{}
}

type Model struct {
	Phases   []*Phase
	Mutex    sync.Mutex
	OnUpdate func()
}

func NewModel() *Model {
	return &Model{
		Phases: make([]*Phase, 0),
	}
}

func (m *Model) startPhase(name string, start time.Time) *Phase {

	if len(m.Phases) > 0 {
		m.Phases[len(m.Phases)-1].Span.SetComplete(start)
	}

	phase := &Phase{
		Name: name,
		Span: Span{
			Start: start,
		},
	}
	m.Phases = append(m.Phases, phase)

	return phase
}

func (m *Model) currentPhase() *Phase {
	if len(m.Phases) > 0 {
		return m.Phases[len(m.Phases)-1]
	}
	return nil
}

func (m *Model) StartJob(phaseName, jobId string, start time.Time) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	phase := m.currentPhase()
	if phase == nil || phaseName != phase.Name {
		phase = m.startPhase(phaseName, start)
	}

	phase.Jobs = append(phase.Jobs, &Job{
		Id:    jobId,
		State: "running",
		Span: Span{
			Start: start,
		},
	})

	if m.OnUpdate != nil {
		m.OnUpdate()
	}
}

func (m *Model) CompleteJob(jobId string, t time.Time) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	phase := m.currentPhase()
	if phase == nil {
		// error
		logger.Error("No current phase")
		return
	}

	job := phase.FindJob(jobId)
	if job == nil {
		logger.Error("No job found", "jobId", jobId)
		return
	}
	job.State = "complete"
	job.Span.SetComplete(t)

	if m.OnUpdate != nil {
		m.OnUpdate()
	}
}

func (m *Model) StartTask(jobId, taskName, subject string, t time.Time) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	phase := m.currentPhase()
	if phase == nil {
		// error
		logger.Error("No current phase")
		return
	}

	job := phase.FindJob(jobId)
	if job == nil {
		logger.Error("No job found")
		return
	}

	job.StartTask(taskName, subject, t)

	if m.OnUpdate != nil {
		m.OnUpdate()
	}
}

func (m *Model) CompleteTask(jobId, taskName string, e error, t time.Time) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	phase := m.currentPhase()
	if phase == nil {
		// error
		logger.Error("No current phase")
		return
	}

	job := phase.FindJob(jobId)
	if job == nil {
		logger.Error("No job found")
		return
	}

	job.CompleteTask(taskName, e, t)

	if m.OnUpdate != nil {
		m.OnUpdate()
	}
}

type Span struct {
	Start    time.Time
	Complete time.Time
}

func (s *Span) SetComplete(t time.Time) {
	s.Complete = t
}

func (s *Span) Duration() time.Duration {
	if s.Complete.IsZero() {
		return time.Since(s.Start)
	}
	return s.Complete.Sub(s.Start)
}

type Phase struct {
	Name string
	Span Span
	Jobs []*Job
}

func (p *Phase) FindJob(jobId string) *Job {
	for _, j := range p.Jobs {
		if j.Id == jobId {
			return j
		}
	}
	return nil
}

type Job struct {
	Id    string
	State string
	Span  Span
	Tasks []*Task
}

func (j *Job) StartTask(taskName, subject string, t time.Time) {
	j.Tasks = append(j.Tasks, &Task{
		Name:    taskName,
		Subject: subject,
		State:   "running",
		Span: Span{
			Start: t,
		},
	})
}

func (j *Job) CompleteTask(taskName string, e error, t time.Time) {
	task := j.FindTask(taskName)
	if task == nil {
		logger.Error("No task found", "taskName", taskName)
		return
	}
	task.State = "complete"
	task.Err = e
	task.Span.SetComplete(t)
}

func (j *Job) FindTask(taskName string) *Task {
	for _, t := range j.Tasks {
		if t.Name == taskName {
			return t
		}
	}
	return nil
}

type Task struct {
	Name    string
	Subject string
	State   string
	Err     error
	Span    Span
}

type TuiView struct {
	Model *Model
}

func NewTuiView(model *Model) *TuiView {
	return &TuiView{
		Model: model,
	}
}

func (v *TuiView) Render() {
	v.Model.Mutex.Lock()
	defer v.Model.Mutex.Unlock()

	var b strings.Builder

	b.WriteString("\x1b[2J\x1b[H")
	b.WriteString("--------------------------------\n")

	for _, phase := range v.Model.Phases {
		fmt.Fprintf(&b, "%s Phase: [%s]\n", phase.Name, phase.Span.Duration())
		for _, job := range phase.Jobs {
			fmt.Fprintf(&b, "  Job: %s [%s] %s\n",
				job.Id,
				job.Span.Duration(),
				stateIcon(job.State))
			for _, task := range job.Tasks {
				errInfo := ""
				if task.Err != nil {
					errInfo = fmt.Sprintf(" (error: %v)", task.Err)
				}
				fmt.Fprintf(&b, "    Task: %s %s [%s] %s%s\n",
					task.Name,
					task.Subject,
					task.Span.Duration(),
					stateIcon(task.State),
					errInfo)
			}
		}
	}

	fmt.Print(b.String())
}

func stateIcon(state string) string {
	switch state {
	case "running":
		return "▶"
	case "complete":
		return "✔"
	default:
		return "?"
	}
}

type TuiBridgeHandler struct {
	Model *Model
}

func NewTuiBridgeHandler(m *Model) *TuiBridgeHandler {
	return &TuiBridgeHandler{
		Model: m,
	}
}

func ignorePhase(phase string) bool {
	return phase == "barrier"
}

func (m *TuiBridgeHandler) Handle(ctx context.Context, r slog.Record) error {

	if r.Level != logger.LevelProgress {
		return nil
	}

	switch r.Message {
	case "JobStart":
		phase, id := getJobInfo(&r)
		if !ignorePhase(phase) {
			m.Model.StartJob(phase, id, r.Time)
		}
	case "JobComplete":
		phase, id := getJobInfo(&r)
		if !ignorePhase(phase) {
			m.Model.CompleteJob(id, r.Time)
		}
	case "TaskStart":
		jobId, taskName, subject, _ := getTaskInfo(&r)
		m.Model.StartTask(jobId, taskName, subject, r.Time)
	case "TaskComplete":
		jobId, taskName, _, err := getTaskInfo(&r)
		m.Model.CompleteTask(jobId, taskName, err, r.Time)
	}

	return nil
}

func getJobInfo(r *slog.Record) (phase, jobId string) {
	job := ""
	r.Attrs(func(a slog.Attr) bool {
		switch a.Key {
		case "job":
			job = a.Value.String()
			return false
		}
		return true
	})
	parts := strings.Split(job, ":")

	phase = parts[0]
	jobId = job

	return phase, jobId
}

func getTaskInfo(r *slog.Record) (jobId, taskName, subject string, err error) {
	_, jobId = getJobInfo(r)

	r.Attrs(func(a slog.Attr) bool {
		switch a.Key {
		case "task":
			taskName = a.Value.String()
		case "subject":
			subject = a.Value.String()
		case "error":
			e := a.Value.Any()
			if e != nil {
				err = e.(error)
			}
		}
		return true
	})

	return jobId, taskName, subject, err
}

func (m *TuiBridgeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewTuiBridgeHandler(m.Model)
}

func (m *TuiBridgeHandler) WithGroup(name string) slog.Handler {
	return NewTuiBridgeHandler(m.Model)
}

func (m *TuiBridgeHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level == logger.LevelProgress
}

type PrettyHandler struct {
}

func NewPrettyHandler() *PrettyHandler {
	return &PrettyHandler{}
}

func (m *PrettyHandler) toString(r *slog.Record) string {
	var b strings.Builder

	fmt.Fprintf(&b, "%s %s", r.Level, r.Message)

	r.Attrs(func(a slog.Attr) bool {
		fmt.Fprintf(&b, " %s=%v", a.Key, a.Value.Any())
		return true
	})

	return b.String()
}

func (m *PrettyHandler) Handle(ctx context.Context, r slog.Record) error {

	switch r.Level {
	case logger.LevelProgress:
		switch r.Message {
		case "JobStart":
			_, id := getJobInfo(&r)
			fmt.Printf("=======================\n")
			fmt.Printf("[Job] Start %q\n", id)
		case "JobComplete":
			_, id := getJobInfo(&r)
			fmt.Printf("[Job] Complete %q\n", id)
		case "TaskStart":
			fmt.Println("-----------------------")
			_, taskName, subject, _ := getTaskInfo(&r)
			fmt.Printf("[Task] Start %q for %q\n", taskName, subject)
		case "TaskComplete":
			_, taskName, subject, err := getTaskInfo(&r)
			fmt.Printf("[Task] Complete %q for %q\n", taskName, subject)
			if err != nil {
				fmt.Printf("  Error: %v\n", err)
			}
		}
	default:
		fmt.Println(m.toString(&r))
	}
	return nil
}

func (m *PrettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewPrettyHandler()
}

func (m *PrettyHandler) WithGroup(name string) slog.Handler {
	return NewPrettyHandler()
}

func (m *PrettyHandler) Enabled(ctx context.Context, level slog.Level) bool {
	switch level {
	case logger.LevelProgress | logger.LevelWarn | logger.LevelError | logger.LevelInfo:
		return true
	default:
		return false
	}
}

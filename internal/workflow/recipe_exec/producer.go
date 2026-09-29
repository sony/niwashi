// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe_exec

import (
	"os"

	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/logger"
	"github.com/sony/niwashi/internal/patch"
	"github.com/sony/niwashi/internal/state"
	"github.com/sony/niwashi/internal/workflow"
	"github.com/sony/niwashi/internal/workflow/applier"
)

type Producer struct {
	workflow.DefaultErrorHolder
	recipeContext       RecipeContext
	task                TaskSpec
	baseActionContext   *baseActionContext
	filteredTasks       []*FilteredTask
	workingTasks        []*Task
	affectSubjects      []*SubjectId
	paramsFileName      string
	patchPathConstraint patch.PathConstraint
}

func NewProducer(ctx RecipeContext, task TaskSpec) (*Producer, error) {

	// separate concurrent tasks
	filteredTasks, affectSubjects, err := FilterTask(ctx, task)
	if err != nil {
		return nil, err
	}

	if len(filteredTasks) == 0 {
		// no tasks to produce
		logger.Info("no available subjects", "task", task.GetName())
		return nil, nil
	}

	baseContext, err := NewBaseActionContext(ctx, task, filteredTasks[0].Adapter)
	if err != nil {
		return nil, err
	}

	// make directories for the task
	ws := baseContext.taskWs
	err = file.MakeDirsAll(
		ctx.GetApplyContext().FileSystem(),
		ws.GetLogDirPath(),
		ws.GetInputDirPath(),
		ws.GetOutputDirPath())
	if err != nil {
		return nil, err
	}

	return &Producer{
		recipeContext:     ctx,
		filteredTasks:     filteredTasks,
		task:              task,
		affectSubjects:    affectSubjects,
		baseActionContext: baseContext,
	}, nil
}

func (p *Producer) StoreFilesForTask() error {

	// Save local files and update workspace info

	fs := p.recipeContext.GetApplyContext().FileSystem()
	//	ws := p.baseActionContext.actionWs
	ws := p.baseActionContext.taskWs

	// params.json
	paramsFileName := file.SanitizeFilename(p.TaskName() + "_params.json")
	paramsFilePath := ws.Join(ws.GetInputDirPath(), paramsFileName)
	params := p.baseActionContext.params
	if err := file.WriteWithEncoding(params, paramsFilePath, fs); err != nil {
		return err
	}

	// state.json
	stateFileName := p.recipeContext.GetStateFileName()
	stateFilePath := ws.Join(ws.GetInputDirPath(), stateFileName)

	if _, err := fs.Stat(stateFilePath); os.IsNotExist(err) {
		runtimeState := p.recipeContext.GetRuntimeState()
		if err := runtimeState.WriteToFile(stateFilePath, fs); err != nil {
			return err
		}
	}

	p.paramsFileName = paramsFileName

	return nil
}

func (p *Producer) GetAffectSubjects() []*SubjectId {
	return p.affectSubjects
}

func (p *Producer) TaskName() string {
	return p.task.GetName()
}

func (p *Producer) GetTask() TaskSpec {
	return p.task
}

func (p *Producer) GetParamsFileName() string {
	return p.paramsFileName
}

func (p *Producer) Produce() *Task {

	if len(p.filteredTasks) == 0 {
		// if there is an error and no working tasks, stop producing tasks and return the error
		return nil
	}

	// generate task from subject

	subTask := p.filteredTasks[0]
	p.filteredTasks = p.filteredTasks[1:]

	// generate task for the subject
	actionContext := p.baseActionContext.NewActionContext(p, subTask)
	t := NewTask(actionContext)

	p.workingTasks = append(p.workingTasks, t)
	return t
}

func (p *Producer) HasWorkingTask() bool {
	return len(p.workingTasks) > 0
}

func (p *Producer) IsComplete() bool {
	return len(p.filteredTasks) == 0 && len(p.workingTasks) == 0
}

func (p *Producer) OnComplete(t *Task, err error) {
	// remove completed task from working tasks
	for i, tsk := range p.workingTasks {
		if tsk == t {
			p.workingTasks = append(p.workingTasks[:i], p.workingTasks[i+1:]...)
			break
		}
	}

	p.HoldError(err)
}

func (p *Producer) createPathConstraint() patch.PathConstraint {

	ctx := p.baseActionContext
	return createPathConstraint(
		ctx.recipeContext.GetJob().GetPhase(),
		ctx.task,
		ctx.taskWs.Target,
	)
}

func (p *Producer) GetPatchPathConstraint() patch.PathConstraint {
	if p.patchPathConstraint == nil {
		p.patchPathConstraint = p.createPathConstraint()
	}
	return p.patchPathConstraint
}

func (p *Producer) Finalize(task *Task) error {

	// close resource used for the task
	err := task.actionContext.Close()
	if err != nil {
		p.HoldError(err)
	}

	// apply patches to the runtime state when the task is succeeded
	if p.HasError() {
		return nil
	}

	tplParams := task.actionContext.tplParams
	constraint := p.GetPatchPathConstraint()

	patches := []patch.Patch{}
	for _, sc := range p.task.GetStateChanges() {
		patch, err := applier.MakeJsonPatch(
			sc, tplParams, p.recipeContext.GetApplyContext().FileSystem(), constraint)
		if err != nil {
			logger.Error("Error creating JSON patch", "error", err)
			p.HoldError(err)
			return err
		}
		patches = append(patches, patch...)
	}

	phase := p.recipeContext.GetJob().GetPhase()
	target := p.baseActionContext.taskWs.Target

	validate := getPatchValidator(phase)

	// apply patches with constraint and validation
	err = p.recipeContext.GetApplyContext().ApplyPatches(patches, constraint, func(s *state.Accessor) error {
		if validate == nil {
			return nil
		}
		return validate(s, target)
	})

	if err != nil {
		logger.Error("Error applying patches", "error", err)
		p.HoldError(err)
	}

	return err
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package action

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/sony/niwashi/internal/clone"
	"github.com/sony/niwashi/internal/types"
)

type TemplateEngine interface {
	Render(input string) (string, error)
}

type Template struct {
	tpl *template.Template
}

func NewTemplate(funcMap template.FuncMap) *Template {
	return &Template{
		tpl: template.New("").Funcs(funcMap),
	}
}

func (t *Template) Render(input string, data any) (string, error) {

	if input == "" {
		return input, nil
	}

	// Replace {{ .Outputs }} with actual values from vars
	i, err := transformOutputs(input)
	if err != nil {
		return "", err
	}

	// Parse and execute the template
	tmpl, err := t.tpl.Parse(i)
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	err = tmpl.Execute(&builder, data)
	if err != nil {
		return "", err
	}

	return builder.String(), nil
}

var re = regexp.MustCompile(`\{\{\s*\.Outputs((?:\.[a-zA-Z0-9_]+)+)\s*\}\}`)

func transformOutputs(text string) (string, error) {

	result := re.ReplaceAllStringFunc(text, func(match string) string {
		submatches := re.FindStringSubmatch(match)
		if len(submatches) != 2 {
			return match
		}
		path := strings.ReplaceAll(submatches[1], ".", "/")
		return "{{ .Outputs }}" + path
	})

	return result, nil
}

func NewFuncMap(data any) template.FuncMap {
	funcMap := template.FuncMap{
		"default": fnDefault,
		"toJson":  fnToJson,
	}

	if data != nil {
		funcMap["render"] = func(s string) (string, error) {
			// except render function
			return NewTemplate(NewFuncMap(nil)).Render(s, data)
		}
	}
	return funcMap
}

func fnToJson(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func fnDefault(def string, val any) string {
	if val == nil {
		return def
	}
	if s, ok := val.(string); ok {
		return s
	}
	return def
}

type defaultEngine struct {
	data *TemplateParams
	tpl  *Template
}

func NewDefaultTemplateEngine(data *TemplateParams) TemplateEngine {
	return &defaultEngine{
		data: data,
		tpl:  NewTemplate(NewFuncMap(data)),
	}
}

func (e *defaultEngine) Render(input string) (string, error) {
	return e.tpl.Render(input, e.data)
}

type TemplateParams struct {
	Params  types.Params
	Store   types.Dict
	Stores  map[string]types.Dict
	Target  string
	Paths   map[string]string
	Outputs string
	Loop    map[string]int
	Runtime types.Dict
}

func NewTemplateParams(ops ...func(*TemplateParams)) *TemplateParams {

	// init
	tp := &TemplateParams{
		Params: make(types.Params),
		Store:  make(types.Dict),
		Stores: make(map[string]types.Dict),
		Paths: map[string]string{
			"cwd": func() string { dir, _ := os.Getwd(); return dir }(),
		},
		Runtime: make(types.Dict),
	}

	tp.Set(ops...)

	return tp
}

func (p *TemplateParams) Clone() *TemplateParams {
	c := &TemplateParams{
		Params:  p.Params.Clone(),
		Store:   p.Store.Clone(),
		Stores:  make(map[string]types.Dict),
		Target:  p.Target,
		Paths:   clone.StringMap(p.Paths),
		Outputs: p.Outputs,
		Runtime: p.Runtime.Clone(),
	}

	for k, v := range p.Stores {
		c.Stores[k] = v.Clone()
	}
	return c
}

func (p *TemplateParams) Set(ops ...func(*TemplateParams)) {
	for _, op := range ops {
		op(p)
	}
}

func WithTaskDirs(ws TaskDirs) func(*TemplateParams) {
	return func(tp *TemplateParams) {
		tp.Outputs = ws.GetOutputDirPath()
		tp.Paths["workspace"] = ws.GetRootDirPath()
		tp.Paths["work_dir"] = ws.GetWorkDirPath()
		tp.Paths["input_dir"] = ws.GetInputDirPath()
		tp.Paths["output_dir"] = ws.GetOutputDirPath()
		tp.Paths["log_dir"] = ws.GetLogDirPath()
	}
}

func WithParams(params map[string]any) func(*TemplateParams) {
	return func(tp *TemplateParams) {
		tp.Params = params
	}
}

func WithStore(store types.Dict) func(*TemplateParams) {
	return func(tp *TemplateParams) {
		tp.Store = store
	}
}

func WithStores(stores map[string]types.Dict) func(*TemplateParams) {
	return func(tp *TemplateParams) {
		tp.Stores = stores
	}
}

func WithRuntime(runtime Runtime) func(*TemplateParams) {
	return func(tp *TemplateParams) {
		tp.Runtime = types.Dict{
			"tool":    runtime.GetTools(),
			"service": runtime.GetService(),
		}
	}
}

func WithTarget(target string) func(*TemplateParams) {
	return func(tp *TemplateParams) {
		tp.Target = target
	}
}

var reAssets = regexp.MustCompile(`\{\{\s*\.Assets\s*\}\}`)

func replaceAssetsDir(tpl string, path string) string {
	return reAssets.ReplaceAllStringFunc(tpl, func(match string) string {
		return path
	})
}

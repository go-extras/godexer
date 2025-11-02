package executor

import (
	"encoding/json"
	"io/fs"
	"strings"

	"github.com/go-extras/errors"
)

func NewIncludeCommand(storage fs.ReadFileFS) func(ectx *ExecutorContext) Command {
	return func(ectx *ExecutorContext) Command {
		return &IncludeCommand{
			SubExecuteCommand: SubExecuteCommand{
				BaseCommand: BaseCommand{
					Ectx: ectx,
				},
			},
			storage: storage,
		}
	}
}

func NewIncludeCommandWithBasePath(storage fs.ReadFileFS, basepath string) func(ectx *ExecutorContext) Command {
	return func(ectx *ExecutorContext) Command {
		return &IncludeCommand{
			SubExecuteCommand: SubExecuteCommand{
				BaseCommand: BaseCommand{
					Ectx: ectx,
				},
			},
			storage:  storage,
			basepath: basepath,
		}
	}
}

type IncludeCommand struct {
	SubExecuteCommand
	// File to include.
	File string `json:"file"`
	// Variables to be passed to the included script.
	Variables map[string]any `json:"variables"`
	// If noMergeVars is true, variables that will be set in the included script
	// will be stored in a newly created "%step_name%_vars" var.
	// Otherwise, they will be placed in the global script namespace.
	NoMergeVars bool `json:"noMergeVars"`

	storage  fs.ReadFileFS
	basepath string
}

func (r *IncludeCommand) Execute(variables map[string]any) error {
	if len(r.File) == 0 {
		return errors.Errorf("filename in %q is empty", r.StepName)
	}

	if r.storage == nil {
		return errors.New("storage is nil")
	}

	filename, ok := MaybeEvalValue(r.File, variables).(string)
	if !ok {
		return errors.Errorf("filename in %q must be a string", r.StepName)
	}
	if r.basepath != "" && filename[0] != '/' {
		filename = strings.TrimRight(r.basepath, "/") + "/" + filename
	}

	script, err := r.storage.ReadFile(strings.TrimPrefix(filename, "/"))
	if err != nil {
		return errors.Wrapf(err, "unable to load included script %q in %q", filename, r.StepName)
	}

	cmds := &RawScenario{}
	script, err = toJSON(script)
	if err != nil {
		return errors.Wrapf(err, "failed to parse input %q in %q", filename, r.StepName)
	}
	if err = json.Unmarshal(script, cmds); err != nil {
		return errors.Wrapf(err, "failed to unmarshal script %q in %q", filename, r.StepName)
	}

	r.RawCommands = cmds.Commands

	var vars map[string]any

	if r.Variables == nil {
		r.Variables = make(map[string]any)
	}

	if !r.NoMergeVars {
		for k, v := range r.Variables {
			variables[k] = v
		}
		vars = variables
		return r.SubExecuteCommand.Execute(vars)
	}

	vars = r.Variables
	vars["_parent"] = variables
	err = r.SubExecuteCommand.Execute(vars)
	delete(vars, "_parent")
	variables[r.GetStepName()+"_variables"] = vars
	return err
}

package executor

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/go-extras/errors"
)

//nolint:gochecknoinits // init is used for automatic command registration
func init() {
	RegisterCommand("foreach", NewForeachCommand)
	RegisterCommand("repeat_for", NewForeachCommand) // deprecated
}

// ForeachCommand allows defining a set of commands that will be run
// for each item in the given map or slice contained in the `variable`.
type ForeachCommand struct {
	BaseCommand
	RawCommands []json.RawMessage `json:"commands"`
	// Value that contains slice or map
	Iterable any `json:"iterable"`
	// Script variable that contains slice or map (unused if iterable is set)
	Variable string `json:"variable"`
	// Script variable that will be created for the key at the iteration (default: key)
	KeyVar string `json:"keyVar"`
	// Script variable that will be created for the value at the iteration (default: value)
	ValueVar string `json:"valueVar"`
	// Script variable that will be created for the parent vars at the iteration (default: parent)
	ParentVar string `json:"parentVar"`

	commands []Command
}

func NewForeachCommand(ectx *ExecutorContext) Command {
	return &ForeachCommand{
		BaseCommand: BaseCommand{
			Ectx: ectx,
		},
	}
}

func (r *ForeachCommand) Execute(variables map[string]any) error {
	if r.Ectx.Executor == nil {
		return errors.Errorf("this command must be run from the executor")
	}

	iterable, err := r.getIterable(variables)
	if err != nil {
		return err
	}

	varMap, varSlice, err := r.convertIterable(iterable)
	if err != nil {
		return err
	}

	if err := r.prepareCommands(); err != nil {
		return err
	}

	return r.executeIterations(varMap, varSlice, variables)
}

func (r *ForeachCommand) getIterable(variables map[string]any) (any, error) {
	if r.Iterable == nil && variables[r.Variable] == nil {
		if r.Variable == "" {
			return nil, errors.Errorf("either iterable or variable must be set")
		}
		return nil, errors.Errorf("variable %q does not exist", r.Variable)
	}

	if r.Iterable != nil {
		return r.Iterable, nil
	}
	return variables[r.Variable], nil
}

func (*ForeachCommand) convertIterable(iterable any) (map[string]any, []any, error) {
	var varMap map[string]any
	var varSlice []any

	switch kind := reflect.TypeOf(iterable).Kind(); kind {
	case reflect.Map:
		varMap = make(map[string]any)
		s := reflect.ValueOf(iterable)
		for _, v := range s.MapKeys() {
			if v.Kind() != reflect.String {
				return nil, nil, errors.Errorf("foreach: invalid map key type %q (expected string)", v.Kind())
			}
			varMap[v.String()] = s.MapIndex(v).Interface()
		}
	case reflect.Slice:
		s := reflect.ValueOf(iterable)
		for i := 0; i < s.Len(); i++ {
			varSlice = append(varSlice, s.Index(i).Interface())
		}
	default:
		return nil, nil, errors.Errorf("foreach: invalid variable type %q (expected slice or map)", kind)
	}

	return varMap, varSlice, nil
}

func (r *ForeachCommand) prepareCommands() error {
	for _, q := range r.RawCommands {
		var tq struct{ Type string }
		if err := json.Unmarshal(q, &tq); err != nil {
			return err
		}
		fn, ok := r.Ectx.Executor.CommandTypeFn(tq.Type)
		if !ok {
			return errors.Errorf("invalid command type: %q", tq.Type)
		}
		cmd := fn(r.Ectx)
		if err := json.Unmarshal(q, cmd); err != nil {
			return err
		}
		r.commands = append(r.commands, cmd)
	}
	return nil
}

func (r *ForeachCommand) executeIterations(varMap map[string]any, varSlice []any, variables map[string]any) error {
	if len(varMap) > 0 {
		for k, v := range varMap {
			if err := foreachSubExecute(r, k, v, variables); err != nil {
				return err
			}
		}
	}

	if len(varSlice) > 0 {
		for k, v := range varSlice {
			if err := foreachSubExecute(r, k, v, variables); err != nil {
				return err
			}
		}
	}

	return nil
}

func foreachSubExecute[T int | string](r *ForeachCommand, k T, v any, variables map[string]any) error {
	ex := r.Ectx.Executor.WithCommands(r.commands, WithStepNameSuffix(fmt.Sprintf("_%v", k)))
	vars := make(map[string]any)
	vars[stringDef(r.ParentVar, "parent")] = variables
	vars[stringDef(r.KeyVar, "key")] = k
	vars[stringDef(r.ValueVar, "value")] = v
	err := ex.Execute(vars)
	if err != nil {
		return err
	}
	return nil
}

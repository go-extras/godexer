package godexer

import (
	"github.com/expr-lang/expr"
	"gopkg.in/Knetic/govaluate.v2"
)

// EvaluatorFunction is the signature for functions available inside `requires:` expressions.
type EvaluatorFunction = func(args ...any) (any, error)

type evaluatorFunctionRegistry map[string]EvaluatorFunction

func newEvaluatorFunctionRegistry() evaluatorFunctionRegistry {
	return make(evaluatorFunctionRegistry)
}

func (r evaluatorFunctionRegistry) clone() map[string]EvaluatorFunction {
	result := make(map[string]EvaluatorFunction, len(r))
	for name, fn := range r {
		result[name] = fn
	}
	return result
}

func (r evaluatorFunctionRegistry) register(name string, fn EvaluatorFunction) {
	r[name] = fn
}

func (r evaluatorFunctionRegistry) registerAll(funcs map[string]EvaluatorFunction) {
	for name, fn := range funcs {
		r.register(name, fn)
	}
}

func (r evaluatorFunctionRegistry) registerLegacy(name string, fn govaluate.ExpressionFunction) {
	r[name] = fn
}

func (r evaluatorFunctionRegistry) registerLegacyAll(funcs map[string]govaluate.ExpressionFunction) {
	for name, fn := range funcs {
		r.registerLegacy(name, fn)
	}
}

func (r evaluatorFunctionRegistry) govaluateFunctions() map[string]govaluate.ExpressionFunction {
	result := make(map[string]govaluate.ExpressionFunction, len(r))
	for name, fn := range r {
		result[name] = fn
	}
	return result
}

func (r evaluatorFunctionRegistry) exprOptions() []expr.Option {
	result := make([]expr.Option, 0, len(r))
	for name, fn := range r {
		result = append(result, expr.Function(name, fn))
	}
	return result
}

// WithRegisteredEvaluatorFunction registers an evaluator function option without exposing third-party types.
func WithRegisteredEvaluatorFunction(name string, fn EvaluatorFunction) func(ex *Executor) {
	return func(ex *Executor) {
		ex.RegisterEvaluatorFunction(name, fn)
	}
}

// WithRegisteredEvaluatorFunctions registers evaluator function options without exposing third-party types.
func WithRegisteredEvaluatorFunctions(funcs map[string]EvaluatorFunction) func(ex *Executor) {
	return func(ex *Executor) {
		ex.RegisterEvaluatorFunctions(funcs)
	}
}

// Deprecated: use WithRegisteredEvaluatorFunction instead.
func WithEvaluatorFunction(name string, fn govaluate.ExpressionFunction) func(ex *Executor) {
	return func(ex *Executor) {
		ex.evaluatorFunctions.registerLegacy(name, fn)
		ex.rebuildGovaluateEvaluatorFunctionCache()
	}
}

// Deprecated: use WithRegisteredEvaluatorFunctions instead.
func WithEvaluatorFunctions(funcs map[string]govaluate.ExpressionFunction) func(ex *Executor) {
	return func(ex *Executor) {
		ex.evaluatorFunctions.registerLegacyAll(funcs)
		ex.rebuildGovaluateEvaluatorFunctionCache()
	}
}

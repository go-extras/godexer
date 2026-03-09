package godexer

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"gopkg.in/Knetic/govaluate.v2"
)

func TestExecutorGovaluateEvaluatorFunctionCacheUpdatesOnRegister(t *testing.T) {
	c := qt.New(t)
	ex := New()

	ex.RegisterEvaluatorFunction("registered", func(args ...any) (any, error) {
		return true, nil
	})
	WithEvaluatorFunction("legacy", govaluate.ExpressionFunction(func(args ...any) (any, error) {
		return true, nil
	}))(ex)

	result, err := ex.evaluateRequiresGovaluate("registered() && legacy()", make(map[string]any))

	c.Assert(err, qt.IsNil)
	c.Assert(result, qt.Equals, true)
	c.Assert(ex.govaluateEvaluatorFunctions, qt.HasLen, 2)
}

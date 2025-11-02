package version

import (
	"github.com/go-extras/errors"
	"github.com/hashicorp/go-version"

	"github.com/go-extras/godexer"
)

var (
	ErrArgMustBeString = errors.New("argument must be string")
	ErrInvalidVersion  = errors.New("invalid version")
)

func parseArgs(args []any) (v1, v2 *version.Version, err error) {
	if len(args) != 2 {
		return nil, nil, errors.New("invalid number of arguments")
	}
	s1, ok := args[0].(string)
	if !ok {
		return nil, nil, errors.Wrap(ErrArgMustBeString, "argument 1 must be string")
	}
	s2, ok := args[1].(string)
	if !ok {
		return nil, nil, errors.Wrap(ErrArgMustBeString, "argument 2 must be string")
	}
	v1, err = version.NewVersion(s1)
	if err != nil {
		return nil, nil, errors.Wrap(errors.WithEquivalents(err, ErrInvalidVersion), "argument 1 must be a valid version")
	}
	v2, err = version.NewVersion(s2)
	if err != nil {
		return nil, nil, errors.Wrap(errors.WithEquivalents(err, ErrInvalidVersion), "argument 2 must be a valid version")
	}
	return v1, v2, nil
}

// WithVersionFuncs registers value evaluator functions.
//
// The following functions are available:
// - `version_lt(version1, version2 string) (bool, error)`
// - `version_lte(version1, version2 string) (bool, error)`
// - `version_gt(version1, version2 string) (bool, error)`
// - `version_gte(version1, version2 string) (bool, error)`
// - `version_eq(version1, version2 string) (bool, error)`
func WithVersionFuncs() func(*executor.Executor) {
	return func(ex *executor.Executor) {
		ex.RegisterEvaluatorFunction("version_lt", func(args ...any) (any, error) {
			v1, v2, err := parseArgs(args)
			if err != nil {
				return false, err
			}
			return v1.LessThan(v2), nil
		})

		ex.RegisterEvaluatorFunction("version_gt", func(args ...any) (any, error) {
			v1, v2, err := parseArgs(args)
			if err != nil {
				return false, err
			}
			return v1.GreaterThan(v2), nil
		})

		ex.RegisterEvaluatorFunction("version_lte", func(args ...any) (any, error) {
			v1, v2, err := parseArgs(args)
			if err != nil {
				return false, err
			}
			return v1.LessThanOrEqual(v2), nil
		})

		ex.RegisterEvaluatorFunction("version_gte", func(args ...any) (any, error) {
			v1, v2, err := parseArgs(args)
			if err != nil {
				return false, err
			}
			return v1.GreaterThanOrEqual(v2), nil
		})

		ex.RegisterEvaluatorFunction("version_eq", func(args ...any) (any, error) {
			v1, v2, err := parseArgs(args)
			if err != nil {
				return false, err
			}
			return v1.Equal(v2), nil
		})
	}
}

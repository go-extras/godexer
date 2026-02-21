package shared_test

import (
	"errors"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/go-extras/godexer/cmd/godexer/shared"
)

func TestExitError(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		c := qt.New(t)
		base := errors.New("something went wrong")
		ee := shared.NewExitError(2, base)
		c.Assert(ee.Error(), qt.Equals, "something went wrong")
	})

	t.Run("Unwrap", func(t *testing.T) {
		c := qt.New(t)
		base := errors.New("wrapped")
		ee := shared.NewExitError(1, base)
		c.Assert(errors.Is(ee, base), qt.IsTrue)
	})

	t.Run("Code", func(t *testing.T) {
		c := qt.New(t)
		ee := shared.NewExitError(3, errors.New("config"))
		c.Assert(ee.Code, qt.Equals, 3)
	})

	t.Run("NewExitErrorf", func(t *testing.T) {
		c := qt.New(t)
		ee := shared.NewExitErrorf(2, "validation failed: %s", "bad type")
		c.Assert(ee.Code, qt.Equals, 2)
		c.Assert(ee.Error(), qt.Equals, "validation failed: bad type")
	})

	t.Run("ErrorsAs", func(t *testing.T) {
		c := qt.New(t)
		ee := shared.NewExitError(1, errors.New("exec"))
		var target *shared.ExitError
		c.Assert(errors.As(ee, &target), qt.IsTrue)
		c.Assert(target.Code, qt.Equals, 1)
	})
}

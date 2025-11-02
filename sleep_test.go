package godexer_test

import (
	"bytes"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
	"github.com/spf13/afero"

	"github.com/go-extras/godexer"
	"github.com/go-extras/godexer/internal/logger"
)

func TestSleepExecute(t *testing.T) {
	c := qt.New(t)
	fs := afero.NewMemMapFs()

	cmd := godexer.NewSleepCommand(&godexer.ExecutorContext{
		Fs:     fs,
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		Logger: &logger.Logger{},
	})
	ex := cmd.(*godexer.SleepCommand)
	ex.Seconds = 10

	godexer.TimeSleep = func(d time.Duration) {
		c.Assert(d.Seconds(), qt.Equals, float64(10))
	}
	defer func() { godexer.TimeSleep = time.Sleep }()
	err := ex.Execute(nil)
	c.Assert(err, qt.IsNil)
}

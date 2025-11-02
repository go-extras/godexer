package executor_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"

	"github.com/go-extras/godexer"
	"github.com/go-extras/godexer/internal/logger"
)

type SleepTestSuite struct {
	suite.Suite
}

func (t *SleepTestSuite) TestExecute() {
	fs := afero.NewMemMapFs()

	cmd := executor.NewSleepCommand(&executor.ExecutorContext{
		Fs:     fs,
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		Logger: &logger.Logger{},
	})
	ex := cmd.(*executor.SleepCommand)
	ex.Seconds = 10

	executor.TimeSleep = func(d time.Duration) {
		t.Equal(float64(10), d.Seconds())
	}
	defer func() { executor.TimeSleep = time.Sleep }()
	err := ex.Execute(nil)
	t.NoError(err)
}

func TestSleep(t *testing.T) {
	suite.Run(t, new(SleepTestSuite))
}

package executor_test

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"

	. "github.com/go-extras/godexer"
	"github.com/go-extras/godexer/internal/logger"
)

type WriteFileTestSuite struct {
	suite.Suite
}

func (t *WriteFileTestSuite) TestExecute() {
	fs := afero.NewMemMapFs()

	cmd := NewWriterFileCommand(&ExecutorContext{
		Fs:     fs,
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	})
	ex := cmd.(*WriteFileCommand)
	ex.File = "dummy"
	ex.Contents = "data {{ index .  \"var1\" }}{{ index .  \"var2\" }}"
	ex.Ectx.Logger = &logger.Logger{}

	m := map[string]any{
		"var1": "val1",
		"var2": "val2",
	}
	err := ex.Execute(m)
	t.NoError(err)

	f, _ := fs.Open("dummy")
	d, _ := ioutil.ReadAll(f)
	t.Equal("data val1val2", string(d))
}

func (t *WriteFileTestSuite) TestExecute_MissingFile() {
	fs := afero.NewMemMapFs()

	cmd := NewWriterFileCommand(&ExecutorContext{
		Fs:     fs,
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	})
	ex := cmd.(*WriteFileCommand)
	ex.Contents = "data {{ index .  \"var1\" }}{{ index .  \"var2\" }}"
	ex.StepName = "step"

	m := map[string]any{
		"var1": "val1",
		"var2": "val2",
	}
	err := ex.Execute(m)
	t.EqualError(err, "filename in \"step\" is empty")
}

func TestWriteFile(t *testing.T) {
	suite.Run(t, new(WriteFileTestSuite))
}

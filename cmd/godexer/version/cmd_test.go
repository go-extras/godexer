package versioncmd_test

import (
	"bytes"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/go-extras/godexer/cmd/godexer/shared"
	versioncmd "github.com/go-extras/godexer/cmd/godexer/version"
)

func TestVersionCmd_Run(t *testing.T) {
	c := qt.New(t)

	cmd := versioncmd.New(&shared.Context{})
	var buf bytes.Buffer
	cmd.Cmd().SetOut(&buf)

	err := cmd.Cmd().Execute()
	c.Assert(err, qt.IsNil)
	c.Assert(buf.String(), qt.Matches, `godexer version .+\n`)
}

func TestVersionCmd_DefaultVersion(t *testing.T) {
	c := qt.New(t)
	c.Assert(versioncmd.AppVersion, qt.Equals, "dev")
}

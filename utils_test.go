package godexer_test

import (
	"strings"
	"testing"

	"al.essio.dev/pkg/shellescape"
	qt "github.com/frankban/quicktest"
)

func escapeArgs(args []string) (result string) {
	if len(args) == 0 {
		return result
	}

	var escaped []string
	for _, s := range args {
		escaped = append(escaped, shellescape.Quote(s))
	}
	return strings.Join(escaped, " ")
}

func TestShellEscape(t *testing.T) {
	c := qt.New(t)
	//v := ShellEscape("NlSvBku3.6c(")
	v := "NlSvBk'u3.6c("
	zz := escapeArgs([]string{"-p" + v})
	c.Assert(v, qt.Equals, "NlSvBk'u3.6c(")
	c.Assert(zz, qt.Equals, "'-pNlSvBk'\"'\"'u3.6c('")
}

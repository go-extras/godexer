package executor_test

import (
	"strings"
	"testing"

	"al.essio.dev/pkg/shellescape"
	"github.com/stretchr/testify/suite"
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

type UtilsTestSuite struct {
	suite.Suite
}

func (t *UtilsTestSuite) TestShellEscape() {
	//v := ShellEscape("NlSvBku3.6c(")
	v := "NlSvBk'u3.6c("
	zz := escapeArgs([]string{"-p" + v})
	t.Require().Equal(v, "NlSvBk'u3.6c(")
	t.Require().Equal(zz, "'-pNlSvBk'\"'\"'u3.6c('")
}

func TestUtils(t *testing.T) {
	suite.Run(t, new(UtilsTestSuite))
}

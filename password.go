package executor

import (
	"crypto/rand"
	"math/big"

	"github.com/go-extras/errors"
)

//nolint:gochecknoinits // init is used for automatic command registration
func init() {
	RegisterCommand("password", NewPassword)
}

const defaultCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randStringRunes(n int, letterRunes []rune, letterRunesLen int) (string, error) {
	b := make([]rune, n)

	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(letterRunesLen)))
		if err != nil {
			return "", err
		}
		b[i] = letterRunes[num.Int64()]
	}

	return string(b), nil
}

type PasswordCommand struct {
	BaseCommand
	Variable string
	Length   int
	Charset  string
}

func NewPassword(ectx *ExecutorContext) Command {
	return &PasswordCommand{
		BaseCommand: BaseCommand{
			Ectx: ectx,
		},
	}
}

func (s *PasswordCommand) Execute(variables map[string]any) error {
	if s.Variable == "" {
		return errors.New("password: variable name cannot be empty")
	}
	if s.Length < 8 {
		s.Length = 8
	}
	if s.Charset == "" {
		s.Charset = defaultCharset
	}

	letterRunes := []rune(s.Charset)
	password, err := randStringRunes(s.Length, letterRunes, len(letterRunes))
	if err != nil {
		return errors.Wrap(err, "failed to generate password")
	}
	variables[s.Variable] = password

	return nil
}

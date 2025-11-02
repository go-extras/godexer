package executor

import (
	"math/rand"
	"time"

	"github.com/go-extras/errors"
)

func init() {
	rand.Seed(time.Now().UnixNano())
	RegisterCommand("password", NewPassword)
}

const defaultCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randStringRunes(n int, letterRunes []rune, letterRunesLen int) string {
	b := make([]rune, n)

	for i := range b {
		b[i] = letterRunes[rand.Intn(letterRunesLen)]
	}

	return string(b)
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
	variables[s.Variable] = randStringRunes(s.Length, letterRunes, len(letterRunes))

	return nil
}

package testutils

import (
	"bytes"

	"github.com/sirupsen/logrus"
)

type SimpleFormatter struct{}

func (f *SimpleFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	msg := entry.Message
	b.WriteString(msg)
	b.WriteByte('\n')

	return b.Bytes(), nil
}

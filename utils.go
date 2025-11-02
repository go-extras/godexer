package executor

import (
	"bytes"
	"io"
	"os"
	"sync"
	"time"
	"unicode"

	"github.com/ghodss/yaml"
	"github.com/spf13/afero"
)

var TimeSleep = time.Sleep

func ShellEscape(cmd string) string {
	result := escapeArgs([]string{cmd})
	return result
}

func fileExists(fs afero.Fs, fname string) (bool, error) {
	if _, err := fs.Stat(fname); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
}

// toJson converts a single YAML document into a JSON document
// or returns an error. If the document appears to be JSON the
// YAML decoding path is not used.
func toJson(data []byte) ([]byte, error) {
	if hasJSONPrefix(data) {
		return data, nil
	}
	return yaml.YAMLToJSON(data)
}

var jsonPrefix = []byte("{")

// hasJSONPrefix returns true if the provided buffer appears to start with
// a JSON open brace.
func hasJSONPrefix(buf []byte) bool {
	return hasPrefix(buf, jsonPrefix)
}

// Return true if the first non-whitespace bytes in buf is prefix.
func hasPrefix(buf []byte, prefix []byte) bool {
	trim := bytes.TrimLeftFunc(buf, unicode.IsSpace)
	return bytes.HasPrefix(trim, prefix)
}

type Buffer struct {
	buf  bytes.Buffer
	lock sync.RWMutex
}

func (b *Buffer) Write(p []byte) (n int, err error) {
	b.lock.Lock()
	defer b.lock.Unlock()
	return b.buf.Write(p)
}

func (b *Buffer) Read(p []byte) (n int, err error) {
	b.lock.RLock()
	defer b.lock.RUnlock()
	return b.buf.Read(p)
}

func (b *Buffer) String() string {
	b.lock.RLock()
	defer b.lock.RUnlock()
	return b.buf.String()
}

type CombinedWriter struct {
	writers []io.Writer
}

func NewCombinedWriter(writers []io.Writer) *CombinedWriter {
	return &CombinedWriter{
		writers: writers,
	}
}

func (w *CombinedWriter) Write(p []byte) (n int, err error) {
	for _, wr := range w.writers {
		_, err := wr.Write(p)
		if err != nil {
			return -1, err
		}
	}
	return len(p), nil
}

func stringDef(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

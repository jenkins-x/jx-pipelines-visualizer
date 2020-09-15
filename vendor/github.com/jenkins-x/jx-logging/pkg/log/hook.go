package log

import (
	"io"
	"os"

	"github.com/pkg/errors"

	"github.com/sirupsen/logrus"
)

// Hook is a logrus hook for splunk
type Hook struct {
	path      string
	levels    []logrus.Level
	formatter logrus.Formatter
}

func NewHook(path string, levels []logrus.Level) *Hook {
	return &Hook{path, levels, &logrus.JSONFormatter{}}
}

func (h *Hook) Fire(entry *logrus.Entry) error {
	b, err := h.formatter.Format(entry)
	if err != nil {
		return err
	}

	err = appendToFile(h.path, string(b))
	if err != nil {
		return err
	}
	return nil
}

// Levels Required for logrus hook implementation
func (h *Hook) Levels() []logrus.Level {
	return h.levels
}

func appendToFile(path, message string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	defer f.Close()

	if err := Append(f, []byte(message)); err != nil {
		return err
	}
	return nil
}

var invalidN bool // initialized to false

func Append(f io.Writer, data []byte) error {
	n, err := f.Write(data)
	if err != nil {
		return errors.Wrapf(err, "failed appending")
	}
	if n != len(data) || invalidN {
		return errors.New("failed appending")
	}
	return nil
}

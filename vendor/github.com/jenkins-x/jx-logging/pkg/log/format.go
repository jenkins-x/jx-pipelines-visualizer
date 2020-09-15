package log

import (
	"bytes"
	"strings"

	"github.com/sirupsen/logrus"
)

// JenkinsXTextFormat lets use a custom text format
type JenkinsXTextFormat struct {
	ShowInfoLevel   bool
	ShowTimestamp   bool
	TimestampFormat string
}

// NewJenkinsXTextFormat creates the default Jenkins X text formatter
func NewJenkinsXTextFormat() *JenkinsXTextFormat {
	return &JenkinsXTextFormat{
		ShowInfoLevel:   false,
		ShowTimestamp:   false,
		TimestampFormat: "2006-01-02 15:04:05",
	}
}

// Format formats the log statement
func (f *JenkinsXTextFormat) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer

	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	level := strings.ToUpper(entry.Level.String())
	switch level {
	case "INFO":
		if f.ShowInfoLevel {
			b.WriteString(colorStatus(level))
			b.WriteString(": ")
		}
	case "WARNING":
		b.WriteString(colorWarn(level))
		b.WriteString(": ")
	case "DEBUG":
		b.WriteString(colorStatus(level))
		b.WriteString(": ")
	default:
		b.WriteString(colorError(level))
		b.WriteString(": ")
	}
	if f.ShowTimestamp {
		b.WriteString(entry.Time.Format(f.TimestampFormat))
		b.WriteString(" - ")
	}

	b.WriteString(entry.Message)

	if !strings.HasSuffix(entry.Message, "\n") {
		b.WriteByte('\n')
	}
	return b.Bytes(), nil
}

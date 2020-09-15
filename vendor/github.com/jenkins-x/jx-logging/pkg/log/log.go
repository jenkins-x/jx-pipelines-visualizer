package log

import (
	"bytes"
	"io"
	"os"

	"github.com/pkg/errors"
	"github.com/rickar/props"

	"github.com/fatih/color"
	stackdriver "github.com/jenkins-x/logrus-stackdriver-formatter/pkg/stackdriver"
	"github.com/sirupsen/logrus"
)

var (
	// colorStatus returns a new function that returns status-colorized (cyan) strings for the
	// given arguments with fmt.Sprint().
	colorStatus = color.New(color.FgCyan).SprintFunc()

	// colorWarn returns a new function that returns status-colorized (yellow) strings for the
	// given arguments with fmt.Sprint().
	colorWarn = color.New(color.FgYellow).SprintFunc()

	// colorError returns a new function that returns error-colorized (red) strings for the
	// given arguments with fmt.Sprint().
	colorError = color.New(color.FgRed).SprintFunc()

	logger *logrus.Entry

	labelsPath = "/etc/labels"
)

// FormatLayoutType the layout kind
type FormatLayoutType string

const (
	// FormatLayoutJSON uses JSON layout
	FormatLayoutJSON FormatLayoutType = "json"

	// FormatLayoutText uses classic colorful Jenkins X layout
	FormatLayoutText FormatLayoutType = "text"

	// FormatLayoutStackdriver uses a custom formatter for stackdriver
	FormatLayoutStackdriver FormatLayoutType = "stackdriver"

	JxLogFormat = "JX_LOG_FORMAT"
	JxLogFile   = "JX_LOG_FILE"
	JxLogLevel  = "JX_LOG_LEVEL"
)

func initializeLogger() error {
	if logger == nil {
		_, err := forceInitLogger()
		if err != nil {
			return err
		}
	}
	return nil
}

func forceInitLogger() (*logrus.Entry, error) {
	// if we are inside a pod, record some useful info
	var fields logrus.Fields

	exists, err := fileExists(labelsPath)
	if err != nil {
		return nil, errors.Wrapf(err, "checking if %s exists", labelsPath)
	}

	if exists {
		f, err := os.Open(labelsPath)
		if err != nil {
			return nil, errors.Wrapf(err, "opening %s", labelsPath)
		}
		labels, err := props.Read(f)
		if err != nil {
			return nil, errors.Wrapf(err, "reading %s as properties", labelsPath)
		}
		app := labels.Get("app")
		if app != "" {
			fields["app"] = app
		}
		chart := labels.Get("chart")
		if chart != "" {
			fields["chart"] = labels.Get("chart")
		}
	}
	logger = logrus.WithFields(fields)

	format := os.Getenv(JxLogFormat)
	if format == "json" {
		setFormatter(FormatLayoutJSON)
	} else if format == "stackdriver" {
		setFormatter(FormatLayoutStackdriver)
	} else {
		setFormatter(FormatLayoutText)
	}

	level := os.Getenv(JxLogLevel)
	if level != "" {
		err := SetLevel(level)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to set level to %s", level)
		}
	}

	debugFile := os.Getenv(JxLogFile)
	if debugFile != "" {
		hook := NewHook(debugFile, logrus.AllLevels)
		logrus.AddHook(hook)
	}

	return logger, nil
}

// Logger obtains the logger for use in the jx codebase
// This is the only way you should obtain a logger
func Logger() *logrus.Entry {
	err := initializeLogger()
	if err != nil {
		logrus.Warnf("error initializing logrus %v", err)
	}
	return logger
}

// SetLevel sets the logging level
func SetLevel(s string) error {
	level, err := logrus.ParseLevel(s)
	if err != nil {
		return errors.Errorf("Invalid log level '%s'", s)
	}
	logrus.SetLevel(level)
	return nil
}

// GetLevel gets the current log level
func GetLevel() string {
	return logrus.GetLevel().String()
}

// GetLevels returns the list of valid log levels
func GetLevels() []string {
	var levels []string
	for _, level := range logrus.AllLevels {
		levels = append(levels, level.String())
	}
	return levels
}

// setFormatter sets the logrus format to use either text or JSON formatting
func setFormatter(layout FormatLayoutType) {
	switch layout {
	case FormatLayoutJSON:
		logrus.SetFormatter(&logrus.JSONFormatter{})
	case FormatLayoutStackdriver:
		logrus.SetFormatter(stackdriver.NewFormatter())
	default:
		logrus.SetFormatter(NewJenkinsXTextFormat())
	}
}

// CaptureOutput calls the specified function capturing and returning all logged messages.
func CaptureOutput(f func()) string {
	var buf bytes.Buffer
	logrus.SetOutput(&buf)
	f()
	logrus.SetOutput(os.Stdout)
	return buf.String()
}

// SetOutput sets the outputs for the default logger.
func SetOutput(out io.Writer) {
	logrus.SetOutput(out)
}

// copied from utils to avoid circular import
func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, errors.Wrapf(err, "failed to check if file exists %s", path)
}

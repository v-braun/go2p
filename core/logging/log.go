package logging

import (
	"github.com/sirupsen/logrus"
)

type Fields logrus.Fields

func NewLogger(name string) *Logger {
	l := logrus.New()
	l.SetLevel(logrus.StandardLogger().Level)
	l.Formatter = logrus.StandardLogger().Formatter
	l.Out = logrus.StandardLogger().Out
	inner := l.WithField("prefix", name)

	result := &Logger{inner: inner}

	return result
}

type Logger struct {
	inner *logrus.Entry
}

// Debug creates a debug message
func (logger *Logger) Debug(fields Fields, args ...interface{}) {
	logger.inner.WithFields(logrus.Fields(fields)).Debug(args...)
}

// Error creates a error message
func (logger *Logger) Error(fields Fields, args ...interface{}) {
	logger.inner.WithFields(logrus.Fields(fields)).Error(args...)
}

// Info creates a info message
func (logger *Logger) Info(fields Fields, args ...interface{}) {
	logger.inner.WithFields(logrus.Fields(fields)).Info(args...)
}

// Warning creates a warning message
func (logger *Logger) Warning(fields Fields, args ...interface{}) {
	logger.inner.WithFields(logrus.Fields(fields)).Warning(args...)
}

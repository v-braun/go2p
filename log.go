package go2p

import (
	"github.com/sirupsen/logrus"
)

func newLogger(name string) *logrus.Entry {
	l := logrus.New()
	l.SetLevel(logrus.StandardLogger().Level)
	l.Formatter = logrus.StandardLogger().Formatter
	l.Out = logrus.StandardLogger().Out
	result := l.WithField("prefix", name)
	return result
}

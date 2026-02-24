package logging

import (
	"os"

	"github.com/sirupsen/logrus"
)

func InitLogger(level logrus.Level) {
	logrus.SetLevel(level)
	logrus.SetOutput(os.Stderr)
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableLevelTruncation: true,
	})
}

func NewLogger(component string) *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(logrus.StandardLogger().Level)
	logger.SetOutput(logrus.StandardLogger().Out)
	logger.SetFormatter(&formatter{
		component:         component,
		internalFormatter: logrus.StandardLogger().Formatter,
	})

	return logger
}

type formatter struct {
	component         string
	internalFormatter logrus.Formatter
}

func (f *formatter) Format(entry *logrus.Entry) ([]byte, error) {
	entry.Data["component"] = f.component
	return f.internalFormatter.Format(entry)
}

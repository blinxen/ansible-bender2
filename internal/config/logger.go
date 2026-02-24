package config

import (
	"github.com/blinxen/ansible-bender2/internal/logging"
	"github.com/sirupsen/logrus"
)

func GetLogger() *logrus.Logger {
	return logging.NewLogger("config")
}

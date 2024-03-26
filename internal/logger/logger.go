package logger

import (
	"github.com/sirupsen/logrus"
)

func New(logLevel string) *logrus.Logger {
	log := logrus.New()
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.InfoLevel
		log.Warn("unknown log level, use info")
	}
	log.SetLevel(level)
	log.SetFormatter(&logrus.JSONFormatter{})

	return log
}

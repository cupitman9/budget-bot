package logger

import (
	"github.com/sirupsen/logrus"
	"os"
)

var log *logrus.Logger

func init() {

	log = logrus.New()

	file, err := os.OpenFile("application.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logrus.Fatal("Не удалось открыть файл логов:", err)
	}

	log.SetOutput(file)

	log.SetLevel(logrus.InfoLevel)

	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
}

func GetLogger() *logrus.Logger {
	return log
}

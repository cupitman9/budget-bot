package logger

import (
	"io"
	"log"
	"os"
)

func SetupLogger(logFile string) {
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Ошибка при создании файла лога:", err)
	}

	log.SetOutput(io.MultiWriter(file, os.Stdout))
}

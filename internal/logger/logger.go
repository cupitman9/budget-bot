package logger

import (
	"io"
	"log"
	"os"
)

var ErrorLogger *log.Logger

func SetupLogger(logFile string) {
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Ошибка при создании файла лога: %v", err)
	}

	ErrorLogger = log.New(io.MultiWriter(file, os.Stderr), "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

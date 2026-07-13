package logger

import (
	"log"
	"os"
)

func New(component string) *log.Logger {
	return log.New(os.Stdout, "["+component+"] ", log.LstdFlags|log.LUTC)
}

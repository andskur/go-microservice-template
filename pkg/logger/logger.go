// Package logger provides a shared logrus-based logger.
package logger

import (
	"sync"

	"github.com/sirupsen/logrus"
)

// Logger represents the shared logger wrapper.
type Logger struct {
	*logrus.Logger
}

var instance *Logger
var once sync.Once

// Log returns a singleton logger instance.
func Log() *Logger {
	once.Do(func() {
		log := logrus.New()
		formatter := &logrus.TextFormatter{
			FullTimestamp: true,
		}
		log.SetFormatter(formatter)

		instance = &Logger{log}
	})
	return instance
}

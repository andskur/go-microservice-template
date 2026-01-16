package logger

import (
	"sync"

	"github.com/sirupsen/logrus"
)

// logger represent standard logger structure
type logger struct {
	*logrus.Logger
}

var instance *logger
var once sync.Once

// Log returns a singleton logger instance
func Log() *logger {
	once.Do(func() {
		log := logrus.New()
		formatter := &logrus.TextFormatter{
			FullTimestamp: true,
		}
		log.SetFormatter(formatter)

		instance = &logger{log}
	})
	return instance
}

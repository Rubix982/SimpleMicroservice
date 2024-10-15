package log

import (
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

func InitLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})
	logger.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)
	return logger
}

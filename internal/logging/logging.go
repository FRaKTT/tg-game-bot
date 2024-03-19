package logging

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

const logFilename = "log"

func LogFile() string {
	return logFilename
}

func SetupLogging() {
	logOutputs := []io.Writer{os.Stdout}

	{
		logFile, err := os.OpenFile(logFilename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			logrus.Errorf("open file %q: %v", logFilename, err)
		} else {
			logOutputs = append(logOutputs, logFile)
		}
	}

	mw := io.MultiWriter(logOutputs...)
	logrus.SetOutput(mw)
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableLevelTruncation: true,
		PadLevelText:           true,
		FullTimestamp:          true,
		TimestampFormat:        "2006-01-02 15:04:05.999",
	})
}

package logger

import (
	"os"

	log "github.com/Sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	log.SetOutput(os.Stdout)

	log.SetLevel(log.DebugLevel)
}

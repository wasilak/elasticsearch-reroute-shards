package libs

import (
	log "github.com/sirupsen/logrus"
	"io"
	"os"
)

type Logger struct {
	Instance  *log.Logger
	LogFormat string
}

func (l *Logger) Init(logFormat string) {
	l.Instance = log.New()
	file, err := os.OpenFile("./reroute_shards.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	mw := io.MultiWriter(os.Stdout, file)
	l.Instance.SetOutput(mw)

	if logFormat == "json" {
		l.Instance.SetFormatter(&log.JSONFormatter{})
	}
}

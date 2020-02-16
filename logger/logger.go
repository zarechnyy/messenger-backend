package logger

import "log"

type Logger struct{}

func (l *Logger) LogErr(err error) {
	if err != nil {
		log.Println(err)
	}
}

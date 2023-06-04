package sherlock

import (
	"io"
)

const (
	DEBUG = Level(1)
	INFO  = Level(2)
	WARN  = Level(3)
	ERROR = Level(4)
	FATAL = Level(5)

	MinLevel = DEBUG
	MaxLevel = FATAL
)

type Logger interface {
	Output(callerSkip int, s string) error
	SetOutput(w io.Writer)
}

type Level int

func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARNING"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	}
	return "INVALID"
}

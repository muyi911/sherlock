package main

import (
	"fmt"
	"io"
	"log"
	"path/filepath"
	"runtime"
	"time"
)

const (
	Host = 1 << iota // 主机名
	Pid              // 进程ID
)

type Sherlock struct {
	Logger
	level  Level
	prefix string
	flag   int
}

type Option func(sherlock *Sherlock)

func WithFileWriter(writer *FileWriter) Option {
	return func(sherlock *Sherlock) {
		writer.level = sherlock.level
		err := writer.Init()
		if err == nil {
			sherlock.SetOutput(writer)
		}
	}
}

func WithConsoleWriter(writer io.Writer) Option {
	return func(sherlock *Sherlock) {
		sherlock.SetOutput(writer)
	}
}

func NewSherlock(level Level, prefix string, flag int, opts ...Option) *Sherlock {
	logger := log.New(io.Discard, prefix, 0)
	s := &Sherlock{
		Logger: logger,
		level:  level,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (l *Sherlock) CheckLevel(level Level) bool {
	if level >= l.level {
		return false
	}
	return true
}

type LoggingSetting struct {
	Dir          string
	Level        int
	Prefix       string
	WriterOption Option
}

func (l *Sherlock) format(f string, args ...interface{}) string {
	logContent := fmt.Sprintf(f, args...)

	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "???"
		line = 0
	} else {
		file = filepath.Base(file)
	}
	caller := fmt.Sprintf("%s:%d", file, line)

	timeStr := time.Now().Format("2006-01-02 15:04:05")

	return fmt.Sprintf("%s %s %s [%s] %s", host, l.level.String(), timeStr, caller, logContent)
}

func (l *Sherlock) DebugF(f string, args ...interface{}) {
	if l.CheckLevel(DEBUG) {
		return
	}
	_ = l.Output(0, l.format(f, args...))
}

func (l *Sherlock) InfoF(f string, args ...interface{}) {
	if l.CheckLevel(INFO) {
		return
	}
	_ = l.Output(0, l.format(f, args...))
}

func (l *Sherlock) WarnF(f string, args ...interface{}) {
	if l.CheckLevel(WARN) {
		return
	}
	_ = l.Output(0, l.format(f, args...))
}

func (l *Sherlock) ErrorF(f string, args ...interface{}) {
	if l.CheckLevel(ERROR) {
		return
	}
	_ = l.Output(0, l.format(f, args...))
}

func (l *Sherlock) FatalF(f string, args ...interface{}) {
	if l.CheckLevel(FATAL) {
		return
	}
	_ = l.Output(0, l.format(f, args...))
}

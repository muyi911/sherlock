package sherlock

import (
	"fmt"
	"io"
	"log"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

//const (
//	Host = 1 << iota // 主机名
//	Pid              // 进程ID
//)

type Sherlock struct {
	level         Level
	pattern       string
	fileLoggers   map[Level]Logger
	consoleLogger Logger
}

type FileWriterSetting struct {
	LogDir        string
	LogName       string
	Level         Level
	MinLevel      Level
	MaxLevel      Level
	MaxSize       uint64 // byte
	MaxFile       uint64 // 最多保留的文件个数
	BufferSize    int    // byte
	FlushInterval int    // second
	CutInterval   int64  // second
}

type Option func(sherlock *Sherlock)

func WithFileWriter(setting *FileWriterSetting) Option {
	return func(sherlock *Sherlock) {
		if setting.MinLevel != 0 || setting.MaxLevel != 0 {
			minLevel := MinLevel
			maxLevel := MaxLevel

			if setting.MinLevel != 0 {
				minLevel = setting.MinLevel
			}

			if setting.MaxLevel != 0 {
				maxLevel = setting.MaxLevel
			}

			if minLevel > maxLevel {
				fmt.Println("MinLevel must be less than MaxLevel")
				return
			}

			for i := minLevel; i <= maxLevel; i++ {
				setting.Level = i
				writer := NewFileWriter(setting)
				err := writer.Init()
				if err != nil {
					fmt.Println(err)
					return
				}
				logger := log.New(writer, "", 0)
				sherlock.fileLoggers[setting.Level] = logger
			}

			return
		}

		if setting.Level == 0 {
			// 如果都没有设置默认DEBUG
			setting.Level = DEBUG
		}

		writer := NewFileWriter(setting)
		err := writer.Init()
		if err != nil {
			fmt.Println(err)
			return
		}

		logger := log.New(writer, "", 0)
		sherlock.fileLoggers[setting.Level] = logger
	}
}

func WithConsoleWriter(writer io.Writer) Option {
	return func(sherlock *Sherlock) {
		logger := log.New(writer, "", 0)
		sherlock.consoleLogger = logger
	}
}

func NewSherlock(level Level, pattern string, opts ...Option) *Sherlock {
	// 默认输出格式
	if pattern == "" {
		pattern = "{host} {level} {time} [{caller}] {msg}"
	}

	s := &Sherlock{
		level:       level,
		pattern:     pattern,
		fileLoggers: make(map[Level]Logger),
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

func (l *Sherlock) format(level Level, f string, args ...interface{}) string {
	logContent := fmt.Sprintf(f, args...)

	_, file, line, ok := runtime.Caller(3)
	if !ok {
		file = "???"
		line = 0
	} else {
		file = filepath.Base(file)
	}
	caller := fmt.Sprintf("%s:%d", file, line)

	timeStr := time.Now().Format("2006-01-02 15:04:05")

	output := strings.ReplaceAll(l.pattern, "{msg}", logContent)
	output = strings.ReplaceAll(output, "{host}", host)
	output = strings.ReplaceAll(output, "{pid}", pidStr)
	output = strings.ReplaceAll(output, "{level}", level.String())
	output = strings.ReplaceAll(output, "{time}", timeStr)
	output = strings.ReplaceAll(output, "{caller}", caller)

	//return fmt.Sprintf("%s %s %s [%s] %s", host, level.String(), timeStr, caller, logContent)
	return output
}

func (l *Sherlock) output(level Level, f string, args ...interface{}) {
	if l.CheckLevel(level) {
		return
	}

	content := l.format(level, f, args...)

	if l.consoleLogger != nil {
		_ = l.consoleLogger.Output(0, content)
	}

	if fileLogger, ok := l.fileLoggers[level]; ok {
		_ = fileLogger.Output(0, content)
	}
}

func (l *Sherlock) DebugF(f string, args ...interface{}) {
	l.output(DEBUG, f, args...)
}

func (l *Sherlock) InfoF(f string, args ...interface{}) {
	l.output(INFO, f, args...)
}

func (l *Sherlock) WarnF(f string, args ...interface{}) {
	l.output(WARN, f, args...)
}

func (l *Sherlock) ErrorF(f string, args ...interface{}) {
	l.output(ERROR, f, args...)
}

func (l *Sherlock) FatalF(f string, args ...interface{}) {
	l.output(FATAL, f, args...)
}

package sherlock

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

type FileWriter struct {
	*bufio.Writer
	mu            sync.Mutex
	file          *os.File
	bytesCounter  uint64 // The number of bytes written to this file
	level         Level
	maxSize       uint64 // byte
	maxFile       uint64
	logDir        string
	logName       string
	bufferSize    int   // byte
	flushInterval int   // second
	cutInterval   int64 // second
}

func NewFileWriter(setting *FileWriterSetting) *FileWriter {
	fw := &FileWriter{
		mu:            sync.Mutex{},
		level:         setting.Level,
		maxSize:       setting.MaxSize,
		maxFile:       setting.MaxFile,
		logDir:        setting.LogDir,
		logName:       setting.LogName,
		bufferSize:    setting.BufferSize,
		flushInterval: setting.FlushInterval,
		cutInterval:   setting.CutInterval,
	}
	if len(fw.logDir) == 0 {
		fw.logDir = "./" // 默认当前目录
	}

	if len(fw.logName) == 0 {
		fw.logName = program // 默认程序名
	}

	if fw.flushInterval == 0 {
		fw.flushInterval = 3 // 默认3秒
	}

	if fw.bufferSize == 0 {
		fw.bufferSize = 4096 // 默认4KB
	}

	if len(fw.logName) == 0 {
		fw.logName = program
	}

	return fw
}

func (fw *FileWriter) Init() (err error) {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	err = fw.rotateFile(0, true)
	if err != nil {
		fmt.Println("rotateFile error:", err)
		return
	}

	go fw.flushLoop()

	if fw.cutInterval > 0 {
		go fw.cutFile()
	}
	return
}

func (fw *FileWriter) Write(p []byte) (n int, err error) {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	if fw.Writer == nil {
		if err = fw.rotateFile(0, false); err != nil {
			return 0, err
		}
	}

	//if fw.bytesCounter+uint64(len(p)) >= fw.maxSize || fw.Writer == nil {
	//	if err := fw.rotateFile(false); err != nil {
	//		fw.exit(err)
	//		return 0, err
	//	}
	//}

	n, err = fw.Writer.Write(p)
	fw.bytesCounter += uint64(n)
	return
}

func (fw *FileWriter) cutFile() {
	for {
		now := time.Now().Unix()
		nextTime := fw.getNextCutTime(now)
		time.Sleep(time.Duration(nextTime-now) * time.Second)
		fw.mu.Lock()
		err := fw.rotateFile(nextTime, false)
		if err != nil {
			fmt.Println("cutFile error:", err)
		}
		fw.deleteOldFile()
		fw.mu.Unlock()
	}
}

func (fw *FileWriter) rotateFile(cutTime int64, isInit bool) (err error) {
	err = fw.flush()
	if err != nil {
		return
	}

	if fw.file != nil {
		err = fw.file.Close()
		if err != nil {
			return
		}
	}

	fw.file, err = fw.createLogFile(cutTime, isInit)
	if err != nil {
		return
	}

	fw.bytesCounter = 0

	fw.Writer = bufio.NewWriterSize(fw.file, fw.bufferSize)
	return
}

func (fw *FileWriter) createLogFile(cutTime int64, isInit bool) (f *os.File, err error) {
	logDir := fw.logDir

	// 生成日志文件夹
	_ = os.Mkdir(logDir, 0777)

	// 判断需要写入的默认日志文件是否存在
	currentLogName := fw.getCurrentLogName()
	currentLogPath := filepath.Join(logDir, currentLogName)
	fileInfo, err := os.Stat(currentLogPath)
	if err != nil {
		if os.IsNotExist(err) {
			f, err = os.Create(currentLogPath)
			if err != nil || isInit {
				return
			}
		} else {
			return
		}
	}

	// 如果文件没内容，直接返回
	fw.bytesCounter = uint64(fileInfo.Size())
	if fw.bytesCounter == 0 {
		f, err = os.OpenFile(currentLogPath, os.O_APPEND|os.O_WRONLY, 0666)
		return
	}

	// 判断是否需要切割
	if isInit {
		modTime := fileInfo.ModTime()
		nextTime := fw.getNextCutTime(modTime.Unix())
		if time.Now().Unix() >= nextTime {
			f, err = fw.createCutFile(currentLogPath, modTime.Unix())
		}
	} else if cutTime > 0 {
		beforeTime := cutTime - fw.cutInterval
		f, err = fw.createCutFile(currentLogPath, beforeTime)
	} else {
		f, err = os.OpenFile(currentLogPath, os.O_APPEND|os.O_WRONLY, 0666)
	}

	return
}

func (fw *FileWriter) createCutFile(currentLogPath string, cutUnix int64) (f *os.File, err error) {
	cutLogName := fw.getCutLogName(time.Unix(cutUnix, 0))
	cutLogPath := filepath.Join(fw.logDir, cutLogName)

	err = os.Rename(currentLogPath, cutLogPath)
	if err != nil {
		return
	}

	f, err = os.Create(currentLogPath)
	if err != nil {
		return
	}

	return
}

func (fw *FileWriter) getNextCutTime(unix int64) int64 {
	return (unix/fw.cutInterval + 1) * fw.cutInterval
}

func (fw *FileWriter) getCurrentLogName() string {
	logName := fw.logName
	logName = strings.Replace(logName, "{level}", strings.ToLower(fw.level.String()), 1)
	return logName
}

func (fw *FileWriter) getCutLogName(t time.Time) string {
	dateFormat := ""

	if fw.cutInterval < 60 {
		dateFormat = "20060102150405" // 秒
	} else if fw.cutInterval < 3600 {
		dateFormat = "200601021504" // 分钟
	} else if fw.cutInterval < 86400 {
		dateFormat = "2006010215" // 小时
	} else {
		dateFormat = "20060102" // 天
	}

	return fmt.Sprintf("%s.%s", fw.getCurrentLogName(), t.Format(dateFormat))
}

func (fw *FileWriter) deleteOldFile() {
	if fw.maxFile <= 0 {
		return
	}

	files, _ := os.ReadDir(fw.logDir)
	ts := time.Now().Unix() - fw.cutInterval*int64(fw.maxFile)
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if find, _ := regexp.MatchString(fw.getCurrentLogName()+".*", file.Name()); find {
			if f, err := file.Info(); err == nil && f.ModTime().Unix() <= ts {
				_ = os.Remove(filepath.Join(fw.logDir, file.Name()))
			}
		}
	}
}

func (fw *FileWriter) flushLoop() {
	for range time.NewTicker(time.Second * time.Duration(fw.flushInterval)).C {
		fw.Sync()
	}
}

func (fw *FileWriter) flush() (err error) {
	if fw.file != nil && fw.Writer != nil {
		err = fw.Writer.Flush()
		if err != nil {
			fmt.Println("flush error:", err)
			return err
		}

		err = fw.file.Sync()
		if err != nil {
			fmt.Println("sync error:", err)
			return err
		}
	}

	return
}

func (fw *FileWriter) Sync() {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	_ = fw.flush()
}

func (fw *FileWriter) exit(err error) {
	_, _ = fmt.Fprintf(os.Stderr, "log exiting error: %s\n", err)
	_ = fw.flush()
	os.Exit(2)
}

func (fw *FileWriter) Close() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	return fw.flush()
}

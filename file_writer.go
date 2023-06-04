package main

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
	file          *os.File
	level         Level
	bytesCounter  uint64 // The number of bytes written to this file
	maxSize       uint64 // byte
	maxFile       uint64
	logDir        string
	logName       string
	bufferSize    int   // byte
	flushInterval int   // second
	cutInterval   int64 // second
	mu            sync.Mutex
}

func NewFileWriter(logDir string, logName string, bufferSize, flushInterval int, cutInterval int64, maxSize uint64, maxFile uint64) *FileWriter {
	fw := &FileWriter{
		maxSize:       maxSize,
		maxFile:       maxFile,
		logDir:        logDir,
		logName:       logName,
		bufferSize:    bufferSize,
		flushInterval: flushInterval,
		cutInterval:   cutInterval,
		mu:            sync.Mutex{},
	}
	if len(logDir) == 0 {
		fw.logDir = "./" // 默认当前目录
	}
	if len(logName) == 0 {
		fw.logName = program // 默认程序名
	}
	if maxSize == 0 {
		fw.maxSize = 1024 * 1024 * 1800 // 默认1800MB
	}
	if flushInterval == 0 {
		fw.flushInterval = 3 // 默认3秒
	}
	if cutInterval == 0 {
		fw.cutInterval = 86400 // 默认24小时 (24 * 60 * 60s)
	}
	if bufferSize == 0 {
		fw.bufferSize = 4096 // 默认4KB
	}
	if len(logName) == 0 {
		fw.logName = program
	}

	go fw.flushLoop()
	go fw.cutFile()
	return fw
}

func (fw *FileWriter) Init() (err error) {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	err = fw.rotateFile(0, true)
	if err != nil {
		fmt.Println("rotateFile error:", err)
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
	if err != nil {
		fw.exit(err)
	}
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

	fw.bytesCounter = 0
	fw.file, err = fw.createLogFile(cutTime, isInit)
	if err != nil {
		return
	}

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

	if fw.file != nil {
		err = fw.file.Close()
		if err != nil {
			return
		}
	}

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
	levelName := fw.level.String()
	return fmt.Sprintf("%s.%s", fw.logName, strings.ToLower(levelName))
}

func (fw *FileWriter) getCutLogName(t time.Time) string {
	return fmt.Sprintf("%s.%04d%02d%02d%02d",
		fw.getCurrentLogName(),
		t.Year(),
		t.Month(),
		t.Day(),
		t.Hour())
}

func (fw *FileWriter) deleteOldFile() {
	files, _ := os.ReadDir(fw.logDir)
	ts := time.Now().Unix() - fw.cutInterval*int64(fw.maxFile)
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if find, _ := regexp.MatchString(fw.getCurrentLogName()+".*", file.Name()); find {
			if f, err := file.Info(); err == nil && f.ModTime().Unix() < ts {
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

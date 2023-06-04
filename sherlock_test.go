package main

import (
	"log"
	"os"
	"testing"
	"time"
)

func Test(t *testing.T) {
	//sherlock := NewSherlock(DEBUG, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile, WithConsoleWriter(os.Stdout))
	sherlock := NewSherlock(DEBUG, "", log.Llongfile, WithConsoleWriter(os.Stdout))
	sherlock.DebugF("测试一下")
}

func TestFileWriter(t *testing.T) {
	fileWriter := NewFileWriter("./logs", "test", 0, 0, 30, 0, 5)
	sherlock := NewSherlock(DEBUG, "", log.Llongfile, WithConsoleWriter(os.Stdout), WithFileWriter(fileWriter))
	ticker := time.NewTicker(time.Second * 1)
	for {
		select {
		case <-ticker.C:
			sherlock.DebugF("测试一下")
		}
	}
}

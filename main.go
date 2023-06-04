package main

import (
	"log"
	"os"
	"time"
)

func main() {
	fileWriter := NewFileWriter("./logs", "test", 4096, 0, 60, 1024*20, 5)
	sherlock := NewSherlock(DEBUG, "", log.Llongfile, WithConsoleWriter(os.Stdout), WithFileWriter(fileWriter))
	ticker := time.NewTicker(time.Second * 1)
	for {
		select {
		case <-ticker.C:
			sherlock.DebugF("测试一下测试一下")
		}
	}
}

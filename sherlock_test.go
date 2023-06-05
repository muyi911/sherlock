package sherlock

import (
	"os"
	"testing"
	"time"
)

func Test(t *testing.T) {

	sherlock := NewSherlock(
		DEBUG,
		WithConsoleWriter(os.Stdout),
		WithFileWriter(&FileWriterSetting{
			LogDir:      "./logs",
			LogName:     "test.{level}",
			Level:       DEBUG,
			MaxLevel:    INFO,
			CutInterval: 10,
			MaxFile:     5,
		}),
	)

	ticker := time.NewTicker(time.Second * 1)
	for {
		select {
		case <-ticker.C:
			sherlock.DebugF("测试一下测试一下")
			sherlock.InfoF("测试一下测试一下")
		}
	}
}

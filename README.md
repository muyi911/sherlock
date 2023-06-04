# Sherlock
自用日志框架

# 功能
-[x] 日志级别 
-[x] 格式化输出（日志前缀）
-[x] 指定输出文件
-[x] 日志切割

# 日志级别
``` go
const (
    DEBUG = Level(1)
    INFO  = Level(2)
    WARN  = Level(3)
    ERROR = Level(4)
    FATAL = Level(5)
)
```

# 日志切割
- 目前只能按照时间维度切割，且必须是小时及以上，不然切割出来的日志会被覆盖
- 切割出来的日志文件在同一目录下，不支持自定义目录
- 后续完善按文件大小切割，以及支持同一时间范围内的多次切割

# 使用方式
使用以下方式引用：
``` go
import (
    "github.com/muyi911/sherlock"
)
```

控制台输出
``` go
sherlock := NewSherlock(
    DEBUG,
    WithConsoleWriter(os.Stdout),
)
```

写入文件，不同日志级别会写入不同文件，不支持将不同的日志文件写入同一个文件，如果设置了`MinLevel`或者`MaxLevel`，则会创建多个fileWriter；如果只设置了`Level`，则只会创建一个fileWriter
``` go
sherlock := NewSherlock(
    DEBUG,
    WithConsoleWriter(os.Stdout),
    WithFileWriter(&FileWriterSetting{
        LogDir:      "./logs",
        LogName:     "test",
        Level:       DEBUG,
        MinLevel:    DEBUG,
        MaxLevel:    FATAL,
        CutInterval: 10,
    }),
)
```
# Sherlock
自用日志框架

# 功能
- [x] 日志级别 
- [x] 格式化输出
- [x] 指定输出文件
- [x] 日志切割

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

# 输出格式化
- 如果pattern为空，默认`{host} {level} {time} [{caller}] {msg}`
- 目前只支持以下参数：

| 格式化参数    | 说明                     |
|----------|------------------------|
| {host}   | 主机名                    |
| {pid}    | 进程id                   |
| {level}  | 日志级别                   |
| {time}   | 时间：yyyy-MM-dd HH:mm:ss |
| {caller} | 日志输出文件和行数              |
| {msg}    | 日志内容                   |


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
    "{host} {level} {time} [{caller}] {msg}",
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

# 输出的文件名称
- 默认会写入`{LogDir}/{LogName}`文件中，如：`./logs/test`
- 为了省略配置，`LogName`支持日志级别的格式化`{Level}`
- 配置`LogName`为：`test.{Level}`，则会根据日志级别写入不同的文件，如：`./logs/test.debug`、`./logs/test.info`、`./logs/test.warn`、`./logs/test.error`、`./logs/test.fatal`
- 暂时只支持`{Level}`，后续会支持更多的格式化参数
- `CutInterval`会影响切割后的文件名称，`0-59`会精确到秒，`60-3599`会精确到分钟，`3600-86399`会精确到小时，`86400-`会精确到天
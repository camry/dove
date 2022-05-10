package log

import "strings"

// Level 是日志级别类型。
type Level int8

const (
    LevelDebug Level = iota - 1 // LevelDebug 是日志调试级别。
    LevelInfo                   // LevelInfo 是日志信息级别。
    LevelWarn                   // LevelWarn 是日志警告级别。
    LevelError                  // LevelError 是日志错误级别。
    LevelFatal                  // LevelFatal 是日志致命级别。
)

// String 返回日志级别对应的字符串。
func (l Level) String() string {
    switch l {
    case LevelDebug:
        return "DEBUG"
    case LevelInfo:
        return "INFO"
    case LevelWarn:
        return "WARN"
    case LevelError:
        return "ERROR"
    case LevelFatal:
        return "FATAL"
    default:
        return ""
    }
}

// ParseLevel 将级别字符串解析为日志级别值。
func ParseLevel(s string) Level {
    switch strings.ToUpper(s) {
    case "DEBUG":
        return LevelDebug
    case "INFO":
        return LevelInfo
    case "WARN":
        return LevelWarn
    case "ERROR":
        return LevelError
    case "FATAL":
        return LevelFatal
    }
    return LevelInfo
}

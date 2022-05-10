package log

import (
    "context"
    "fmt"
    "os"
)

// DefaultMessageKey 默认消息KEY。
var DefaultMessageKey = "msg"

// Option 是日志助手选项类型。
type Option func(*Helper)

// Helper 是一个日志助手。
type Helper struct {
    logger Logger
    msgKey string
}

// WithMessageKey 配置消息KEY。
func WithMessageKey(k string) Option {
    return func(opts *Helper) {
        opts.msgKey = k
    }
}

// NewHelper 新建一个日志助手。
func NewHelper(logger Logger, opts ...Option) *Helper {
    options := &Helper{
        msgKey: DefaultMessageKey,
        logger: logger,
    }
    for _, o := range opts {
        o(options)
    }
    return options
}

// WithContext 返回被 h 的浅副本改变的上下文 ctx, 提供的 ctx 不能为空。
func (h *Helper) WithContext(ctx context.Context) *Helper {
    return &Helper{
        msgKey: h.msgKey,
        logger: WithContext(ctx, h.logger),
    }
}

// Log 按级别和键值打印日志。
func (h *Helper) Log(level Level, keyvals ...any) {
    _ = h.logger.Log(level, keyvals...)
}

// Debug 打印调试级别的日志。
func (h *Helper) Debug(a ...any) {
    h.Log(LevelDebug, h.msgKey, fmt.Sprint(a...))
}

// Debugf 按 fmt.Sprintf 格式打印调试级别的日志。
func (h *Helper) Debugf(format string, a ...any) {
    h.Log(LevelDebug, h.msgKey, fmt.Sprintf(format, a...))
}

// Debugw 按键值对打印调试级别的日志。
func (h *Helper) Debugw(keyvals ...any) {
    h.Log(LevelDebug, keyvals...)
}

// Info 打印信息级别的日志。
func (h *Helper) Info(a ...any) {
    h.Log(LevelInfo, h.msgKey, fmt.Sprint(a...))
}

// Infof 按 fmt.Sprintf 格式打印信息级别的日志。
func (h *Helper) Infof(format string, a ...any) {
    h.Log(LevelInfo, h.msgKey, fmt.Sprintf(format, a...))
}

// Infow 按键值对打印信息级别的日志。
func (h *Helper) Infow(keyvals ...any) {
    h.Log(LevelInfo, keyvals...)
}

// Warn 打印警告级别的日志。
func (h *Helper) Warn(a ...any) {
    h.Log(LevelWarn, h.msgKey, fmt.Sprint(a...))
}

// Warnf 按 fmt.Sprintf 格式打印警告级别的日志。
func (h *Helper) Warnf(format string, a ...any) {
    h.Log(LevelWarn, h.msgKey, fmt.Sprintf(format, a...))
}

// Warnw 按键值对打印警告级别的日志。
func (h *Helper) Warnw(keyvals ...any) {
    h.Log(LevelWarn, keyvals...)
}

// Error 打印错误级别的日志。
func (h *Helper) Error(a ...any) {
    h.Log(LevelError, h.msgKey, fmt.Sprint(a...))
}

// Errorf 按 fmt.Sprintf 格式打印错误级别的日志。
func (h *Helper) Errorf(format string, a ...any) {
    h.Log(LevelError, h.msgKey, fmt.Sprintf(format, a...))
}

// Errorw 按键值对打印错误级别的日志。
func (h *Helper) Errorw(keyvals ...any) {
    h.Log(LevelError, keyvals...)
}

// Fatal 打印致命级别的日志。
func (h *Helper) Fatal(a ...any) {
    h.Log(LevelFatal, h.msgKey, fmt.Sprint(a...))
    os.Exit(1)
}

// Fatalf 按 fmt.Sprintf 格式打印致命级别的日志。
func (h *Helper) Fatalf(format string, a ...any) {
    h.Log(LevelFatal, h.msgKey, fmt.Sprintf(format, a...))
    os.Exit(1)
}

// Fatalw 按键值对打印致命级别的日志。
func (h *Helper) Fatalw(keyvals ...any) {
    h.Log(LevelFatal, keyvals...)
    os.Exit(1)
}

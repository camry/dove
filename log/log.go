package log

import (
    "context"
    "log"
)

// DefaultLogger 是默认的日志记录器。
var DefaultLogger Logger = NewStdLogger(log.Writer())

// Logger 是一个日志接口。
type Logger interface {
    Log(level Level, keyvals ...any) error
}

type logger struct {
    logs      []Logger
    prefix    []any
    hasValuer bool
    ctx       context.Context
}

// Log 按级别和键值打印日志。
func (c *logger) Log(level Level, keyvals ...any) error {
    kvs := make([]any, 0, len(c.prefix)+len(keyvals))
    kvs = append(kvs, c.prefix...)
    if c.hasValuer {
        bindValues(c.ctx, kvs)
    }
    kvs = append(kvs, keyvals...)
    for _, l := range c.logs {
        if err := l.Log(level, kvs...); err != nil {
            return err
        }
    }
    return nil
}

// With 配置日志字段。
func With(l Logger, kv ...any) Logger {
    if c, ok := l.(*logger); ok {
        kvs := make([]any, 0, len(c.prefix)+len(kv))
        kvs = append(kvs, kv...)
        kvs = append(kvs, c.prefix...)
        return &logger{
            logs:      c.logs,
            prefix:    kvs,
            hasValuer: containsValuer(kvs),
            ctx:       c.ctx,
        }
    }
    return &logger{logs: []Logger{l}, prefix: kv, hasValuer: containsValuer(kv)}
}

// WithContext 返回被 l 的浅副本改变的上下文 ctx, 提供的 ctx 不能为空。
func WithContext(ctx context.Context, l Logger) Logger {
    if c, ok := l.(*logger); ok {
        return &logger{
            logs:      c.logs,
            prefix:    c.prefix,
            hasValuer: c.hasValuer,
            ctx:       ctx,
        }
    }
    return &logger{logs: []Logger{l}, ctx: ctx}
}

// MultiLogger 包装多个日志记录器。
func MultiLogger(logs ...Logger) Logger {
    return &logger{logs: logs}
}

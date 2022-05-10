package log

import (
    "bytes"
    "fmt"
    "io"
    "log"
    "sync"
)

var _ Logger = (*stdLogger)(nil)

type stdLogger struct {
    log  *log.Logger
    pool *sync.Pool
}

// NewStdLogger 新建一个日志记录器。
func NewStdLogger(w io.Writer) Logger {
    return &stdLogger{
        log: log.New(w, "", 0),
        pool: &sync.Pool{
            New: func() any {
                return new(bytes.Buffer)
            },
        },
    }
}

// Log 打印键值对日志。
func (l *stdLogger) Log(level Level, keyvals ...any) error {
    if len(keyvals) == 0 {
        return nil
    }
    if (len(keyvals) & 1) == 1 {
        keyvals = append(keyvals, "KEYVALS UNPAIRED")
    }
    buf := l.pool.Get().(*bytes.Buffer)
    buf.WriteString(level.String())
    for i := 0; i < len(keyvals); i += 2 {
        _, _ = fmt.Fprintf(buf, " %s=%v", keyvals[i], keyvals[i+1])
    }
    _ = l.log.Output(4, buf.String())
    buf.Reset()
    l.pool.Put(buf)
    return nil
}

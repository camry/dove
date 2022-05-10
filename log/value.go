package log

import (
    "context"
    "runtime"
    "strconv"
    "strings"
    "time"
)

var (
    defaultDepth = 3
    // DefaultCaller 是一个返回文件和行号的 Valuer。
    DefaultCaller = Caller(defaultDepth)

    // DefaultTimestamp 是一个返回当前时间的 Valuer。
    DefaultTimestamp = Timestamp(time.RFC3339)
)

// Valuer 是返回一个日志值。
type Valuer func(ctx context.Context) any

// Value 返回函数值。
func Value(ctx context.Context, v any) any {
    if v, ok := v.(Valuer); ok {
        return v(ctx)
    }
    return v
}

// Caller 返回一个调用者的 pkg/file:line 描述的 Valuer。
func Caller(depth int) Valuer {
    return func(context.Context) any {
        _, file, line, _ := runtime.Caller(depth)
        if strings.LastIndex(file, "/log/filter.go") > 0 {
            depth++
            _, file, line, _ = runtime.Caller(depth)
        }
        if strings.LastIndex(file, "/log/helper.go") > 0 {
            depth++
            _, file, line, _ = runtime.Caller(depth)
        }
        idx := strings.LastIndexByte(file, '/')
        return file[idx+1:] + ":" + strconv.Itoa(line)
    }
}

// Timestamp 返回一个自定义时间格式的 Valuer
func Timestamp(layout string) Valuer {
    return func(context.Context) any {
        return time.Now().Format(layout)
    }
}

// bindValues 绑定 Valuer 类型的值。
func bindValues(ctx context.Context, keyvals []any) {
    for i := 1; i < len(keyvals); i += 2 {
        if v, ok := keyvals[i].(Valuer); ok {
            keyvals[i] = v(ctx)
        }
    }
}

// containsValuer 是否包含 Valuer 类型的值。
func containsValuer(keyvals []any) bool {
    for i := 1; i < len(keyvals); i += 2 {
        if _, ok := keyvals[i].(Valuer); ok {
            return true
        }
    }
    return false
}

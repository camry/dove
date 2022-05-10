package log

// FilterOption 是过滤器选项。
type FilterOption func(*Filter)

const fuzzyStr = "***"

// FilterLevel 配置过滤器级别。
func FilterLevel(level Level) FilterOption {
    return func(opts *Filter) {
        opts.level = level
    }
}

// FilterKey 配置过滤器键。
func FilterKey(key ...string) FilterOption {
    return func(o *Filter) {
        for _, v := range key {
            o.key[v] = struct{}{}
        }
    }
}

// FilterValue 配置过滤器值。
func FilterValue(value ...string) FilterOption {
    return func(o *Filter) {
        for _, v := range value {
            o.value[v] = struct{}{}
        }
    }
}

// FilterFunc 配置过滤器方法。
func FilterFunc(f func(level Level, keyvals ...any) bool) FilterOption {
    return func(o *Filter) {
        o.filter = f
    }
}

// Filter 是一个日志过滤器。
type Filter struct {
    logger Logger
    level  Level
    key    map[any]struct{}
    value  map[any]struct{}
    filter func(level Level, keyvals ...any) bool
}

// NewFilter 新建一个日志过滤器。
func NewFilter(logger Logger, opts ...FilterOption) *Filter {
    options := Filter{
        logger: logger,
        key:    make(map[any]struct{}),
        value:  make(map[any]struct{}),
    }
    for _, o := range opts {
        o(&options)
    }
    return &options
}

// Log 按级别和键值打印日志。
func (f *Filter) Log(level Level, keyvals ...any) error {
    if level < f.level {
        return nil
    }
    // fkv 用于提供一个切片来包含过滤器的前缀和键值对。
    var fkv []any
    if l, ok := f.logger.(*logger); ok {
        if len(l.prefix) > 0 {
            fkv = make([]any, 0, len(l.prefix)+len(keyvals))
            fkv = append(fkv, l.prefix...)
            fkv = append(fkv, keyvals...)
        }
    } else {
        fkv = keyvals
    }
    if f.filter != nil && f.filter(level, fkv...) {
        return nil
    }
    if len(f.key) > 0 || len(f.value) > 0 {
        for i := 0; i < len(keyvals); i += 2 {
            v := i + 1
            if v >= len(keyvals) {
                continue
            }
            if _, ok := f.key[keyvals[i]]; ok {
                keyvals[v] = fuzzyStr
            }
            if _, ok := f.value[keyvals[v]]; ok {
                keyvals[v] = fuzzyStr
            }
        }
    }
    return f.logger.Log(level, keyvals...)
}

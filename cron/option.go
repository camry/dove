package cron

import (
    "time"

    "github.com/camry/dove/log"
)

// Option 表示对 Cron 的默认行为的修改。
type Option func(*Cron)

// WithLocation 覆盖 cron 实例的时区。
func WithLocation(loc *time.Location) Option {
    return func(c *Cron) {
        c.location = loc
    }
}

// WithSeconds 覆盖用于解析任务计划的解析器以包含秒字段作为第一个字段。
func WithSeconds() Option {
    return WithParser(NewParser(
        Second | Minute | Hour | Dom | Month | Dow | Descriptor,
    ))
}

// WithParser 覆盖用于解析任务计划的解析器。
func WithParser(p ScheduleParser) Option {
    return func(c *Cron) {
        c.parser = p
    }
}

// WithChain 指定要应用于添加到此 cron 的所有任务的包装器。
// 有关提供的包装器，请参阅此包中的 Chain* 函数。
func WithChain(wrappers ...JobWrapper) Option {
    return func(c *Cron) {
        c.chain = NewChain(wrappers...)
    }
}

// WithLogger 使用提供的日志记录器。
func WithLogger(logger *log.Helper) Option {
    return func(c *Cron) {
        c.logger = logger
    }
}

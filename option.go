package dove

import (
    "context"
    "os"
    "time"

    "github.com/camry/dove/log"
    "github.com/camry/dove/server"
)

// Option 定义一个应用程序选项类型。
type Option func(o *option)

// option 应用程序选项实体对象。
type option struct {
    id      string
    name    string
    version string

    ctx     context.Context
    signals []os.Signal

    logger      *log.Helper
    stopTimeout time.Duration
    servers     []server.Server
}

// ID 配置服务ID。
func ID(id string) Option {
    return func(o *option) { o.id = id }
}

// Name 配置服务名称。
func Name(name string) Option {
    return func(o *option) { o.name = name }
}

// Version 配置服务版本。
func Version(version string) Option {
    return func(o *option) { o.version = version }
}

// Context 配置服务上下文。
func Context(ctx context.Context) Option {
    return func(o *option) { o.ctx = ctx }
}

// Signals 配置服务信号。
func Signals(signals ...os.Signal) Option {
    return func(o *option) { o.signals = signals }
}

// Logger 配置日志记录器。
func Logger(logger log.Logger) Option {
    return func(o *option) { o.logger = log.NewHelper(logger) }
}

// StopTimeout 配置应用停止超时时间（单位：秒）。
func StopTimeout(t time.Duration) Option {
    return func(o *option) { o.stopTimeout = t }
}

// Server 配置服务器。
func Server(srv ...server.Server) Option {
    return func(o *option) { o.servers = srv }
}

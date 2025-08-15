package dove

import (
    "context"
    "os"
    "time"

    "github.com/camry/g/glog"

    "github.com/camry/dove/server"
)

// Option 定义一个应用程序选项类型。
type Option func(o *option)

// option 应用程序选项实体对象。
type option struct {
    id      string
    name    string
    version string

    ctx  context.Context
    sigs []os.Signal

    logger      glog.Logger
    stopTimeout time.Duration
    servers     []server.Server

    // Before and After funcs
    beforeStart []func(context.Context) error
    beforeStop  []func(context.Context) error
    afterStart  []func(context.Context) error
    afterStop   []func(context.Context) error
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

// Signal 配置服务信号。
func Signal(signals ...os.Signal) Option {
    return func(o *option) { o.sigs = signals }
}

// Logger 配置日志记录器。
func Logger(logger glog.Logger) Option {
    return func(o *option) { o.logger = logger }
}

// StopTimeout 配置应用停止超时时间（单位：秒）。
func StopTimeout(t time.Duration) Option {
    return func(o *option) { o.stopTimeout = t }
}

// Server 配置服务器。
func Server(srv ...server.Server) Option {
    return func(o *option) { o.servers = srv }
}

/**********************************/
/******** Before and After ********/
/**********************************/

// BeforeStart 应用启动前执行此 funcs。
func BeforeStart(fn func(context.Context) error) Option {
    return func(o *option) {
        o.beforeStart = append(o.beforeStart, fn)
    }
}

// BeforeStop 应用停止前执行此 funcs。
func BeforeStop(fn func(context.Context) error) Option {
    return func(o *option) {
        o.beforeStop = append(o.beforeStop, fn)
    }
}

// AfterStart 应用启动后执行此 funcs。
func AfterStart(fn func(context.Context) error) Option {
    return func(o *option) {
        o.afterStart = append(o.afterStart, fn)
    }
}

// AfterStop 应用停止后执行此 funcs。
func AfterStop(fn func(context.Context) error) Option {
    return func(o *option) {
        o.afterStop = append(o.afterStop, fn)
    }
}

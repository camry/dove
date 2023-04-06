package dove

import (
    "context"
    "errors"
    "os"
    "os/signal"
    "sync"
    "syscall"
    "time"

    "github.com/camry/g/glog"
    "github.com/google/uuid"
    "golang.org/x/sync/errgroup"
)

// AppInfo 应用程序上下文值接口。
type AppInfo interface {
    ID() string
    Name() string
    Version() string
}

// App 应用程序组件生命周期管理器。
type App struct {
    opt    option
    ctx    context.Context
    cancel func()
}

// New 创建应用生命周期管理器。
func New(opts ...Option) *App {
    o := option{
        ctx:         context.Background(),
        sigs:        []os.Signal{syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT},
        stopTimeout: 10 * time.Second,
    }
    if id, err := uuid.NewUUID(); err == nil {
        o.id = id.String()
    }
    for _, opt := range opts {
        opt(&o)
    }
    if o.logger != nil {
        glog.SetLogger(o.logger)
    }
    ctx, cancel := context.WithCancel(o.ctx)
    return &App{
        ctx:    ctx,
        cancel: cancel,
        opt:    o,
    }
}

// ID 返回服务实例ID。
func (a *App) ID() string { return a.opt.id }

// Name 返回服务名称。
func (a *App) Name() string { return a.opt.name }

// Version 返回服务版本号。
func (a *App) Version() string { return a.opt.version }

// Run 执行应用程序生命周期中注册的所有服务。
func (a *App) Run() (err error) {
    sctx := NewContext(a.ctx, a)
    eg, ctx := errgroup.WithContext(sctx)
    wg := sync.WaitGroup{}

    for _, fn := range a.opt.beforeStart {
        if err = fn(sctx); err != nil {
            return err
        }
    }

    // 启动注册的服务器。
    for _, srv := range a.opt.servers {
        srv := srv
        eg.Go(func() error {
            <-ctx.Done() // 等待停止信号
            stopCtx, cancel := context.WithTimeout(sctx, a.opt.stopTimeout)
            defer cancel()
            return srv.Stop(stopCtx)
        })
        wg.Add(1)
        eg.Go(func() error {
            wg.Done()
            return srv.Start(sctx)
        })
    }
    wg.Wait()

    for _, fn := range a.opt.afterStart {
        if err = fn(sctx); err != nil {
            return err
        }
    }

    c := make(chan os.Signal, 1)
    signal.Notify(c, a.opt.sigs...)
    // 停止应用程序。
    eg.Go(func() error {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-c:
            return a.Stop()
        }
    })
    if err = eg.Wait(); err != nil && !errors.Is(err, context.Canceled) {
        return err
    }
    for _, fn := range a.opt.afterStop {
        err = fn(sctx)
    }
    return nil
}

// Stop 优雅的停止应用程序。
func (a *App) Stop() (err error) {
    sctx := NewContext(a.ctx, a)
    for _, fn := range a.opt.beforeStop {
        err = fn(sctx)
    }
    if a.cancel != nil {
        a.cancel()
    }
    return nil
}

type appKey struct{}

// NewContext 返回一个带有值的新上下文。
func NewContext(ctx context.Context, a AppInfo) context.Context {
    return context.WithValue(ctx, appKey{}, a)
}

// FromContext 返回存储在 ctx 中的传输值（如果有）。
func FromContext(ctx context.Context) (a AppInfo, ok bool) {
    a, ok = ctx.Value(appKey{}).(AppInfo)
    return
}

package dove

import (
    "context"
    "errors"
    "os"
    "os/signal"
    "sync"
    "syscall"
    "time"

    "github.com/google/uuid"
    "golang.org/x/sync/errgroup"
)

type App struct {
    opt    *option
    ctx    context.Context
    cancel func()
}

// New 创建应用生命周期管理器。
func New(opts ...Option) *App {
    o := &option{
        ctx:         context.Background(),
        signals:     []os.Signal{syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGTERM},
        stopTimeout: 10 * time.Second,
    }
    if id, err := uuid.NewUUID(); err != nil {
        o.id = id.String()
    }
    for _, opt := range opts {
        opt(o)
    }
    ctx, cancel := context.WithCancel(o.ctx)
    return &App{
        opt:    o,
        ctx:    ctx,
        cancel: cancel,
    }
}

// ID 返回服务实例ID。
func (a *App) ID() string { return a.opt.id }

// Name 返回服务名称。
func (a *App) Name() string { return a.opt.name }

// Version 返回服务版本号。
func (a *App) Version() string { return a.opt.version }

// Run 执行应用程序生命周期中注册的所有服务。
func (a *App) Run() error {
    eg, ctx := errgroup.WithContext(a.ctx)
    wg := sync.WaitGroup{}
    c := make(chan os.Signal, 1)
    signal.Notify(c, a.opt.signals...)

    // 启动注册的网络服务器。
    for _, srv := range a.opt.servers {
        srv := srv
        eg.Go(func() error {
            <-ctx.Done() // 等待停止信号
            stopCtx, cancel := context.WithTimeout(a.opt.ctx, a.opt.stopTimeout)
            defer cancel()
            return srv.Stop(stopCtx)
        })
        wg.Add(1)
        eg.Go(func() error {
            wg.Done()
            return srv.Start(a.opt.ctx)
        })
    }
    wg.Wait()

    // 停止应用程序。
    eg.Go(func() error {
        for {
            select {
            case <-ctx.Done():
                return ctx.Err()
            case <-c:
                if err := a.Stop(); err != nil {
                    return err
                }
            }
        }
    })
    if err := eg.Wait(); err != nil && !errors.Is(err, context.Canceled) {
        return err
    }
    return nil
}

// Stop 优雅的停止应用程序。
func (a *App) Stop() error {
    if a.cancel != nil {
        a.cancel()
    }
    return nil
}

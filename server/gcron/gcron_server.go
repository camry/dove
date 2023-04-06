package gcron

import (
    "context"
    "github.com/camry/g/glog"

    cron "github.com/camry/g/gcron"
)

// ServerOption 定义一个 Cron 服务选项类型。
type ServerOption func(s *Server)

type Server struct {
    *cron.Cron

    cronOpts []cron.Option
}

// Options 配置 Cron 选项。
func Options(cronOpts ...cron.Option) ServerOption {
    return func(s *Server) { s.cronOpts = cronOpts }
}

// NewServer 新建 Cron 服务器。
func NewServer(opts ...ServerOption) *Server {
    srv := &Server{}
    for _, opt := range opts {
        opt(srv)
    }
    srv.Cron = cron.New(srv.cronOpts...)
    return srv
}

// Start 启动 Cron 服务。
func (s *Server) Start(ctx context.Context) error {
    glog.Info("[CRON] server starting")
    s.Cron.Start()
    return nil
}

// Stop 停止 Cron 服务。
func (s *Server) Stop(ctx context.Context) error {
    glog.Info("[CRON] server stopping")
    s.Cron.Stop(ctx)
    return nil
}

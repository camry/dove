package gcron

import (
    "context"

    "github.com/camry/dove/cron"
    "github.com/camry/dove/log"
)

// ServerOption 定义一个 Cron 服务选项类型。
type ServerOption func(s *Server)

type Server struct {
    *cron.Cron
    log      *log.Helper
    cronOpts []cron.Option
}

// Options 配置 gRPC 选项。
func Options(cronOpts ...cron.Option) ServerOption {
    return func(s *Server) { s.cronOpts = cronOpts }
}

// NewServer 新建 Cron 服务器。
func NewServer(opts ...ServerOption) *Server {
    srv := &Server{
        log: log.NewHelper(log.GetLogger()),
    }
    for _, opt := range opts {
        opt(srv)
    }
    srv.Cron = cron.New(srv.cronOpts...)
    return srv
}

// Start 启动 Cron 服务。
func (s *Server) Start(ctx context.Context) error {
    s.log.Info("[CRON] server starting")
    s.Cron.Start()
    return nil
}

// Stop 停止 Cron 服务。
func (s *Server) Stop(ctx context.Context) error {
    s.log.Info("[CRON] server stopping")
    s.Cron.Stop(ctx)
    return nil
}

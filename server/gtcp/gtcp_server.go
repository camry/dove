package gtcp

import (
    "context"
    "crypto/tls"
    "github.com/camry/g/glog"
    "github.com/camry/g/gnet/gtcp"
)

// ServerOption 定义一个 TCP 服务选项类型。
type ServerOption func(s *Server)

// Logger 配置日志记录器。
func Logger(logger glog.Logger) ServerOption {
    return func(s *Server) { s.log = glog.NewHelper(logger) }
}

// Address 配置服务监听地址。
func Address(address string) ServerOption {
    return func(s *Server) { s.address = address }
}

// TLSConfig 配置 TLS。
func TLSConfig(c *tls.Config) ServerOption {
    return func(s *Server) { s.tlsConfig = c }
}

// Handler 配置处理器。
func Handler(handler func(conn *gtcp.Conn)) ServerOption {
    return func(s *Server) { s.handler = handler }
}

// Server 定义 TCP 服务包装器。
type Server struct {
    *gtcp.Server

    address   string           // 服务器监听地址。
    handler   func(*gtcp.Conn) // 连接处理器。
    tlsConfig *tls.Config      // TLS 配置。
    log       *glog.Helper     // 日志助手
}

// NewServer 新建 TCP 服务器。
func NewServer(opts ...ServerOption) *Server {
    srv := &Server{
        log: glog.NewHelper(glog.GetLogger()),
    }
    for _, opt := range opts {
        opt(srv)
    }
    srv.Server = gtcp.NewServer(srv.address, srv.handler, srv.tlsConfig)
    return srv
}

// Start 启动 TCP 服务器。
func (s *Server) Start(ctx context.Context) (err error) {
    s.log.Infof("[TCP] server listening on: %s", s.Listener().Addr().String())
    return s.Run(ctx)
}

// Stop 停止 TCP 服务器。
func (s *Server) Stop(ctx context.Context) error {
    s.log.Info("[TCP] server stopping")
    return s.Close(ctx)
}

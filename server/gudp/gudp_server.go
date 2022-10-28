package gudp

import (
    "context"
    "github.com/camry/g/glog"
    "github.com/camry/g/gnet/gudp"
)

// ServerOption 定义一个 UDP 服务选项类型。
type ServerOption func(s *Server)

// Logger 配置日志记录器。
func Logger(logger glog.Logger) ServerOption {
    return func(s *Server) { s.log = glog.NewHelper(logger) }
}

// Address 配置服务监听地址。
func Address(address string) ServerOption {
    return func(s *Server) { s.address = address }
}

// Handler 配置处理器。
func Handler(handler func(*gudp.Conn)) ServerOption {
    return func(s *Server) { s.handler = handler }
}

// Server 定义 UDP 服务器。
type Server struct {
    *gudp.Server

    network string           // UDP 服务器网络协议。
    address string           // UDP 服务器监听地址。
    handler func(*gudp.Conn) // UDP 连接的处理程序。
    log     *glog.Helper     // 日志助手。
}

// NewServer 新建 UDP 服务器。
func NewServer(opts ...ServerOption) *Server {
    srv := &Server{
        network: "udp",
        address: ":0",
        handler: func(conn *gudp.Conn) {},
        log:     glog.NewHelper(glog.GetLogger()),
    }
    for _, opt := range opts {
        opt(srv)
    }
    srv.Server = gudp.NewServer(srv.address, srv.handler)
    return srv
}

// Start 启动 UDP 服务器。
func (s *Server) Start(ctx context.Context) error {
    s.log.Infof("[UDP] server listening on: %s", s.Conn().LocalAddr().String())
    return s.Run(ctx)
}

// Stop 停止 UDP 服务器。
func (s *Server) Stop(ctx context.Context) error {
    s.log.Info("[UDP] server stopping")
    return s.Close(ctx)
}

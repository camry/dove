package gudp

import (
    "context"

    "github.com/camry/g/v2/glog"
    "github.com/camry/g/v2/gnet/gudp"
)

// ServerOption 定义一个 UDP 服务选项类型。
type ServerOption func(s *Server)

// Address 配置服务监听地址。
func Address(address string) ServerOption {
    return func(s *Server) { s.address = address }
}

// Handler 配置处理器。
func Handler(handler func(conn *gudp.ServerConn)) ServerOption {
    return func(s *Server) { s.handler = handler }
}

// Server 定义 UDP 服务器。
type Server struct {
    *gudp.Server

    address string                      // UDP 服务器监听地址。
    handler func(conn *gudp.ServerConn) // UDP 连接的处理程序。
}

// NewServer 新建 UDP 服务器。
func NewServer(opts ...ServerOption) *Server {
    srv := &Server{
        address: ":0",
        handler: func(conn *gudp.ServerConn) {},
    }
    for _, opt := range opts {
        opt(srv)
    }
    srv.Server = gudp.NewServer(srv.address, srv.handler)
    return srv
}

// Start 启动 UDP 服务器。
func (s *Server) Start(ctx context.Context) error {
    glog.Infof("[UDP] server listening on %s", s.GetListenedAddress())
    return s.Run(ctx)
}

// Stop 停止 UDP 服务器。
func (s *Server) Stop(ctx context.Context) error {
    glog.Info("[UDP] server stopping")
    return s.Close(ctx)
}

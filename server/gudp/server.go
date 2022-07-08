package gudp

import (
    "context"
    "github.com/camry/dove/log"
    "net"
)

// ServerOption 定义一个 UDP 服务选项类型。
type ServerOption func(s *Server)

// Logger 配置日志记录器。
func Logger(logger log.Logger) ServerOption {
    return func(s *Server) { s.log = log.NewHelper(logger) }
}

// Address 配置服务监听地址。
func Address(address string) ServerOption {
    return func(s *Server) { s.address = address }
}

// Handler 配置处理器。
func Handler(handler func(*Conn)) ServerOption {
    return func(s *Server) { s.handler = handler }
}

// Server 定义 UDP 服务器。
type Server struct {
    conn    *Conn       // UDP 服务器连接对象
    network string      // UDP 服务器网络协议。
    address string      // UDP 服务器监听地址
    handler func(*Conn) // UDP 连接的处理程序。
    log     *log.Helper // 日志助手。
}

// NewServer 新建 UDP 服务器。
func NewServer(opts ...ServerOption) *Server {
    srv := &Server{
        network: "udp",
        address: ":0",
        handler: func(conn *Conn) {},
        log:     log.NewHelper(log.GetLogger()),
    }
    for _, opt := range opts {
        opt(srv)
    }
    return srv
}

// Start 启动 UDP 服务器。
func (s *Server) Start(ctx context.Context) error {
    addr, err := net.ResolveUDPAddr(s.network, s.address)
    if err != nil {
        return err
    }
    conn, err := net.ListenUDP(s.network, addr)
    if err != nil {
        return err
    }
    s.conn = NewConnByNetConn(conn)
    s.log.Infof("[UDP] server listening on: %s", s.conn.LocalAddr().String())
    s.handler(s.conn)
    return nil
}

// Stop 停止 UDP 服务器。
func (s *Server) Stop(ctx context.Context) error {
    s.log.Info("[UDP] server stopping")
    return s.conn.Close()
}

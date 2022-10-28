package gtcp

import (
    "context"
    "crypto/tls"
    "errors"
    "fmt"
    "net"
    "sync"

    "github.com/camry/g/glog"
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
func Handler(handler func(*Conn)) ServerOption {
    return func(s *Server) { s.handler = handler }
}

// Server 定义 TCP 服务包装器。
type Server struct {
    mu        sync.Mutex   // 用于 Server.listen 并发安全。
    listen    net.Listener // 网络监听器。
    network   string       // 服务器网络协议。
    address   string       // 服务器监听地址。
    handler   func(*Conn)  // 连接处理器。
    tlsConfig *tls.Config  // TLS 配置。
    log       *glog.Helper // 日志助手
}

// NewServer 新建 TCP 服务器。
func NewServer(opts ...ServerOption) *Server {
    srv := &Server{
        network: "tcp",
        address: ":0",
        handler: func(conn *Conn) {},
        log:     glog.NewHelper(glog.GetLogger()),
    }
    for _, opt := range opts {
        opt(srv)
    }
    return srv
}

// Start 启动 TCP 服务器。
func (s *Server) Start(ctx context.Context) (err error) {
    if s.handler == nil {
        err = errors.New("start running failed: socket handler not defined")
        return
    }
    if s.tlsConfig != nil {
        s.mu.Lock()
        s.listen, err = tls.Listen(s.network, s.address, s.tlsConfig)
        s.mu.Unlock()
        if err != nil {
            err = fmt.Errorf(`tls.Listen failed for address "%s"`, s.address)
            return
        }
    } else {
        var tcpAddr *net.TCPAddr
        if tcpAddr, err = net.ResolveTCPAddr(s.network, s.address); err != nil {
            err = fmt.Errorf(`net.ResolveTCPAddr failed for address "%s"`, s.address)
            return err
        }
        s.mu.Lock()
        s.listen, err = net.ListenTCP(s.network, tcpAddr)
        s.mu.Unlock()
        if err != nil {
            err = fmt.Errorf(`net.ListenTCP failed for address "%s"`, s.address)
            return err
        }
    }
    s.log.Infof("[TCP] server listening on: %s", s.listen.Addr().String())
    for {
        var conn net.Conn
        if conn, err = s.listen.Accept(); err != nil {
            err = fmt.Errorf(`Listener.Accept failed`)
            return err
        } else if conn != nil {
            go s.handler(NewConnByNetConn(conn))
        }
    }
}

// Stop 停止 TCP 服务器。
func (s *Server) Stop(ctx context.Context) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    if s.listen == nil {
        return nil
    }
    s.log.Info("[TCP] server stopping")
    return s.listen.Close()
}

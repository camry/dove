package ghttp

import (
    "context"
    "crypto/tls"
    "errors"
    "net"
    "net/http"

    "github.com/camry/g/glog"
)

// ServerOption 定义一个 HTTP 服务选项类型。
type ServerOption func(s *Server)

// Server 定义 HTTP 服务包装器。
type Server struct {
    *http.Server
    log     *glog.Helper
    err     error
    network string
    address string
    tlsConf *tls.Config
    lis     net.Listener
    handler http.Handler
}

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
    return func(s *Server) { s.tlsConf = c }
}

// Handler 配置处理器。
func Handler(handler http.Handler) ServerOption {
    return func(s *Server) { s.handler = handler }
}

// NewServer 新建 HTTP 服务器。
func NewServer(opts ...ServerOption) *Server {
    srv := &Server{
        network: "tcp",
        address: ":0",
        log:     glog.NewHelper(glog.GetLogger()),
    }
    for _, opt := range opts {
        opt(srv)
    }
    srv.Server = &http.Server{
        Handler:   srv.handler,
        TLSConfig: srv.tlsConf,
    }
    srv.err = srv.listen()
    return srv
}

// Start 启动 HTTP 服务。
func (s *Server) Start(ctx context.Context) error {
    if s.err != nil {
        return s.err
    }
    s.BaseContext = func(net.Listener) context.Context {
        return ctx
    }
    s.log.Infof("[HTTP] server listening on: %s", s.lis.Addr().String())
    var err error
    if s.tlsConf != nil {
        err = s.ServeTLS(s.lis, "", "")
    } else {
        err = s.Serve(s.lis)
    }
    if !errors.Is(err, http.ErrServerClosed) {
        return err
    }
    return nil
}

// Stop 停止 HTTP 服务。
func (s *Server) Stop(ctx context.Context) error {
    s.log.Info("[HTTP] server stopping")
    return s.Shutdown(ctx)
}

// listen 网络监听。
func (s *Server) listen() error {
    lis, err := net.Listen(s.network, s.address)
    if err != nil {
        return err
    }
    s.lis = lis
    return nil
}

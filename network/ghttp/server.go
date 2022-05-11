package ghttp

import (
    "context"
    "github.com/camry/dove/log"
    "github.com/labstack/echo-contrib/pprof"
    "github.com/labstack/echo/v4"
    "net"
)

// ServerOption 定义一个 HTTP 服务选项类型。
type ServerOption func(s *Server)

// Server 定义 HTTP 服务包装器。
type Server struct {
    *echo.Echo
    addr        string
    certFile    any
    keyFile     any
    enablePProf bool
    log         *log.Helper
}

// Addr 配置服务地址。
func Addr(addr string) ServerOption {
    return func(s *Server) { s.addr = addr }
}

// TlsFile 配置 HTTPS 服务证书文件。
func TlsFile(certFile, keyFile any) ServerOption {
    return func(s *Server) {
        s.certFile = certFile
        s.keyFile = keyFile
    }
}

// EnablePProf 配置启用 PProf 性能分析工具。
func EnablePProf() ServerOption {
    return func(s *Server) { s.enablePProf = true }
}

// NewServer 新建 HTTP 服务器。
func NewServer(opts ...ServerOption) *Server {
    srv := &Server{
        Echo:     echo.New(),
        addr:     ":0",
        certFile: "",
        keyFile:  "",
        log:      log.NewHelper(log.GetLogger()),
    }
    for _, opt := range opts {
        opt(srv)
    }
    if srv.enablePProf {
        pprof.Register(srv.Echo)
    }
    srv.HideBanner = true
    return srv
}

// Start 启动 HTTP 服务。
func (s *Server) Start(ctx context.Context) error {
    s.Echo.Server.BaseContext = func(net.Listener) context.Context {
        return ctx
    }
    s.log.Info("HTTP server starting")
    if s.certFile != "" && s.keyFile != "" {
        return s.Echo.StartTLS(s.addr, s.certFile, s.keyFile)
    }
    return s.Echo.Start(s.addr)
}

// Stop 停止 HTTP 服务。
func (s *Server) Stop(ctx context.Context) error {
    s.log.Info("HTTP server stopping")
    return s.Shutdown(ctx)
}

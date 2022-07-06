package grpc

import (
    "context"
    "crypto/tls"
    "net"
    "time"

    "github.com/camry/dove/log"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials"
    "google.golang.org/grpc/health"
    "google.golang.org/grpc/health/grpc_health_v1"
    "google.golang.org/grpc/reflection"
)

type ServerOption func(s *Server)

// Logger 配置日志记录器。
func Logger(logger log.Logger) ServerOption {
    return func(s *Server) { s.log = log.NewHelper(logger) }
}

// Address 配置服务监听地址。
func Address(address string) ServerOption {
    return func(s *Server) { s.address = address }
}

// Timeout 配置服务超时时间（单位：秒）。
func Timeout(timeout time.Duration) ServerOption {
    return func(s *Server) { s.timeout = timeout }
}

// TLSConfig 配置 TLS。
func TLSConfig(c *tls.Config) ServerOption {
    return func(s *Server) { s.tlsConf = c }
}

// UnaryInterceptor 配置一元拦截器。
func UnaryInterceptor(in ...grpc.UnaryServerInterceptor) ServerOption {
    return func(s *Server) { s.unaryInterceptors = in }
}

// StreamInterceptor 配置流拦截器。
func StreamInterceptor(in ...grpc.StreamServerInterceptor) ServerOption {
    return func(s *Server) { s.streamInterceptors = in }
}

// Options 配置 gRPC 选项。
func Options(grpcOpts ...grpc.ServerOption) ServerOption {
    return func(s *Server) { s.grpcOpts = grpcOpts }
}

type Server struct {
    *grpc.Server
    baseCtx            context.Context
    log                *log.Helper
    err                error
    network            string
    address            string
    timeout            time.Duration
    tlsConf            *tls.Config
    lis                net.Listener
    grpcOpts           []grpc.ServerOption
    unaryInterceptors  []grpc.UnaryServerInterceptor
    streamInterceptors []grpc.StreamServerInterceptor
    health             *health.Server
}

// NewServer 新建 gRPC 服务器。
func NewServer(opts ...ServerOption) *Server {
    srv := &Server{
        baseCtx: context.Background(),
        network: "tcp",
        address: ":0",
        timeout: 1 * time.Second,
        health:  health.NewServer(),
        log:     log.NewHelper(log.GetLogger()),
    }
    for _, o := range opts {
        o(srv)
    }
    unaryInterceptors := []grpc.UnaryServerInterceptor{
        srv.defaultUnaryServerInterceptor(),
    }
    streamInterceptors := []grpc.StreamServerInterceptor{
        srv.defaultStreamServerInterceptor(),
    }
    if len(srv.unaryInterceptors) > 0 {
        unaryInterceptors = append(unaryInterceptors, srv.unaryInterceptors...)
    }
    if len(srv.streamInterceptors) > 0 {
        streamInterceptors = append(streamInterceptors, srv.streamInterceptors...)
    }
    grpcOpts := []grpc.ServerOption{
        grpc.ChainUnaryInterceptor(unaryInterceptors...),
        grpc.ChainStreamInterceptor(streamInterceptors...),
    }
    if srv.tlsConf != nil {
        grpcOpts = append(grpcOpts, grpc.Creds(credentials.NewTLS(srv.tlsConf)))
    }
    if len(srv.grpcOpts) > 0 {
        grpcOpts = append(grpcOpts, srv.grpcOpts...)
    }
    srv.Server = grpc.NewServer(grpcOpts...)
    srv.err = srv.listen()
    // 内部注册
    grpc_health_v1.RegisterHealthServer(srv.Server, srv.health)
    reflection.Register(srv.Server)
    return srv
}

// Start 启动 gRPC 服务器。
func (s *Server) Start(ctx context.Context) error {
    if s.err != nil {
        return s.err
    }
    s.baseCtx = ctx
    s.log.Infof("[gRPC] server listening on: %s", s.lis.Addr().String())
    s.health.Resume()
    return s.Serve(s.lis)
}

// Stop 停止 gRPC 服务器。
func (s *Server) Stop(ctx context.Context) error {
    s.health.Shutdown()
    s.GracefulStop()
    s.log.Info("[gRPC] server stopping")
    return nil
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

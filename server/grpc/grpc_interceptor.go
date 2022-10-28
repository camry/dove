package grpc

import (
    "context"

    ic "github.com/camry/dove/internal/context"
    "google.golang.org/grpc"
)

// unaryServerInterceptor 默认 gRPC 一元拦截器。
func (s *Server) defaultUnaryServerInterceptor() grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
        ctx, cancel := ic.Merge(ctx, s.baseCtx)
        defer cancel()
        if s.timeout > 0 {
            ctx, cancel = context.WithTimeout(ctx, s.timeout)
            defer cancel()
        }
        h := func(ctx context.Context, req interface{}) (interface{}, error) {
            return handler(ctx, req)
        }
        reply, err := h(ctx, req)
        return reply, err
    }
}

// streamServerInterceptor 默认 gRPC 流拦截器。
func (s *Server) defaultStreamServerInterceptor() grpc.StreamServerInterceptor {
    return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
        ctx, cancel := ic.Merge(ss.Context(), s.baseCtx)
        defer cancel()
        ws := NewWrappedStream(ctx, ss)
        err := handler(srv, ws)
        return err
    }
}

// wrappedStream 重写 gRPC 流上下文。
type wrappedStream struct {
    grpc.ServerStream
    ctx context.Context
}

func NewWrappedStream(ctx context.Context, stream grpc.ServerStream) grpc.ServerStream {
    return &wrappedStream{
        ServerStream: stream,
        ctx:          ctx,
    }
}

func (w *wrappedStream) Context() context.Context {
    return w.ctx
}

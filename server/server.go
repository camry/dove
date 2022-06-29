package network

import "context"

// Server 定义服务接口。
type Server interface {
    Start(context.Context) error
    Stop(context.Context) error
}

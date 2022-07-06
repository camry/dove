package gudp

import (
    "net"
    "time"
)

// Conn 处理 UDP 连接。
type Conn struct {
    *net.UDPConn                    // 底层 UDP 连接。
    remoteAddr        *net.UDPAddr  // 远程地址
    receiveDeadline   time.Time     // 读取数据的超时点。
    sendDeadline      time.Time     // 写入数据的超时点。
    receiveBufferWait time.Duration // 读取缓冲区的间隔持续时间。
}

const (
    defaultRetryInterval  = 100 * time.Millisecond // 重试间隔。
    defaultReadBufferSize = 1024                   // （字节）缓冲区大小。
    receiveAllWaitTimeout = time.Millisecond       // 读取缓冲区的默认间隔。
)

type Retry struct {
    Count    int           // 最大重试次数。
    Interval time.Duration // 重试间隔。
}

// NewConnByNetConn 使用指定的 *net.UDPConn 对象创建一个 UDP 连接对象。
func NewConnByNetConn(conn *net.UDPConn) *Conn {
    return &Conn{
        UDPConn:           conn,
        receiveDeadline:   time.Time{},
        sendDeadline:      time.Time{},
        receiveBufferWait: receiveAllWaitTimeout,
    }
}

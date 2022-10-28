package gtcp

import (
    "crypto/tls"
    "fmt"
    "net"
    "time"
)

const (
    defaultConnTimeout    = 30 * time.Second       // 默认连接超时。
    defaultRetryInternal  = 100 * time.Millisecond // 默认重试间隔。
    defaultReadBufferSize = 128                    // （字节）用于读取的缓冲区大小。
)

type Retry struct {
    Count    int           // 重试计数。
    Interval time.Duration // 重试间隔。
}

// NewNetConn 创建并返回具有指定地址的 net.Conn，例如“127.0.0.1:80”。
// 可选参数`timeout`指定拨号连接的超时时间。
func NewNetConn(address string, timeout ...time.Duration) (net.Conn, error) {
    var (
        network  = `tcp`
        duration = defaultConnTimeout
    )
    if len(timeout) > 0 {
        duration = timeout[0]
    }
    conn, err := net.DialTimeout(network, address, duration)
    if err != nil {
        err = fmt.Errorf(
            `net.DialTimeout failed with network "%s", address "%s", timeout "%s"`,
            network, address, duration,
        )
    }
    return conn, err
}

// NewNetConnTLS 创建并返回具有指定地址的 TLS net.Conn，例如“127.0.0.1:80”。
// 可选参数`timeout`指定拨号连接的超时时间。
func NewNetConnTLS(address string, tlsConfig *tls.Config, timeout ...time.Duration) (net.Conn, error) {
    var (
        network = `tcp`
        dialer  = &net.Dialer{
            Timeout: defaultConnTimeout,
        }
    )
    if len(timeout) > 0 {
        dialer.Timeout = timeout[0]
    }
    conn, err := tls.DialWithDialer(dialer, network, address, tlsConfig)
    if err != nil {
        err = fmt.Errorf(
            `tls.DialWithDialer failed with network "%s", address "%s", timeout "%s", tlsConfig "%v"`,
            network, address, dialer.Timeout, tlsConfig,
        )
    }
    return conn, err
}

// Send 创建到 `address` 的连接，将 `data` 写入连接，然后关闭连接。
// 可选参数 `retry` 指定写入数据失败时的重试策略。
func Send(address string, data []byte, retry ...Retry) error {
    conn, err := NewConn(address)
    if err != nil {
        return err
    }
    defer conn.Close()
    return conn.Send(data, retry...)
}

// SendReceive 创建到 `address` 的连接，将 `data` 写入连接，接收响应，然后关闭连接。
//
// 参数 `length` 指定等待接收的字节数。 它接收所有缓冲区内容并在 `length` 为 -1 时返回。
// 可选参数 `retry` 指定写入数据失败时的重试策略。
func SendReceive(address string, data []byte, length int, retry ...Retry) ([]byte, error) {
    conn, err := NewConn(address)
    if err != nil {
        return nil, err
    }
    defer conn.Close()
    return conn.SendReceive(data, length, retry...)
}

// SendWithTimeout 发送具有写入超时限制的逻辑。
func SendWithTimeout(address string, data []byte, timeout time.Duration, retry ...Retry) error {
    conn, err := NewConn(address)
    if err != nil {
        return err
    }
    defer conn.Close()
    return conn.SendWithTimeout(data, timeout, retry...)
}

// SendReceiveWithTimeout 执行具有读取超时限制的 SendReceive 逻辑。
func SendReceiveWithTimeout(address string, data []byte, receive int, timeout time.Duration, retry ...Retry) ([]byte, error) {
    conn, err := NewConn(address)
    if err != nil {
        return nil, err
    }
    defer conn.Close()
    return conn.SendReceiveWithTimeout(data, receive, timeout, retry...)
}

// isTimeout 检查给定的 `err` 是否是超时错误。
func isTimeout(err error) bool {
    if err == nil {
        return false
    }
    if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
        return true
    }
    return false
}

// MustGetFreePort 执行 GetFreePort，但发生任何错误都会 panic。
func MustGetFreePort() int {
    port, err := GetFreePort()
    if err != nil {
        panic(err)
    }
    return port
}

// GetFreePort 检索并返回一个空闲的端口。
func GetFreePort() (port int, err error) {
    var (
        network = `tcp`
        address = `:0`
    )
    resolvedAddr, err := net.ResolveTCPAddr(network, address)
    if err != nil {
        return 0, fmt.Errorf(
            `net.ResolveTCPAddr failed for network "%s", address "%s"`,
            network, address,
        )
    }
    l, err := net.ListenTCP(network, resolvedAddr)
    if err != nil {
        return 0, fmt.Errorf(
            `net.ListenTCP failed for network "%s", address "%s"`,
            network, resolvedAddr.String(),
        )
    }
    port = l.Addr().(*net.TCPAddr).Port
    if err = l.Close(); err != nil {
        err = fmt.Errorf(
            `close listening failed for network "%s", address "%s", port "%d"`,
            network, resolvedAddr.String(), port,
        )
    }
    return
}

// GetFreePorts 检索并返回指定数量的空闲端口。
func GetFreePorts(count int) (ports []int, err error) {
    var (
        network = `tcp`
        address = `:0`
    )
    for i := 0; i < count; i++ {
        resolvedAddr, err := net.ResolveTCPAddr(network, address)
        if err != nil {
            return nil, fmt.Errorf(
                `net.ResolveTCPAddr failed for network "%s", address "%s"`,
                network, address,
            )
        }
        l, err := net.ListenTCP(network, resolvedAddr)
        if err != nil {
            return nil, fmt.Errorf(
                `net.ListenTCP failed for network "%s", address "%s"`,
                network, resolvedAddr.String(),
            )
        }
        ports = append(ports, l.Addr().(*net.TCPAddr).Port)
        _ = l.Close()
    }
    return ports, nil
}

package gudp

import (
    "errors"
    "fmt"
    "io"
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

// NewConn 创建到 `remoteAddress` 的 UDP 连接。
// 可选参数“localAddress”指定连接的本地地址。
func NewConn(remoteAddress string, localAddress ...string) (*Conn, error) {
    if conn, err := NewNetConn(remoteAddress, localAddress...); err == nil {
        return NewConnByNetConn(conn), nil
    } else {
        return nil, err
    }
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

// Send 将数据写入远程地址。
func (c *Conn) Send(data []byte, retry ...Retry) (err error) {
    for {
        if c.remoteAddr != nil {
            _, err = c.WriteToUDP(data, c.remoteAddr)
        } else {
            _, err = c.Write(data)
        }
        if err != nil {
            // 连接关闭。
            if err == io.EOF {
                return err
            }
            // 即使重试后仍然失败。
            if len(retry) == 0 || retry[0].Count == 0 {
                return errors.New("write data failed")
            }
            if len(retry) > 0 {
                retry[0].Count--
                if retry[0].Interval == 0 {
                    retry[0].Interval = defaultRetryInterval
                }
                time.Sleep(retry[0].Interval)
            }
        } else {
            return nil
        }
    }
}

// Receive 从远程地址接收和返回数据。
func (c *Conn) Receive(buffer int, retry ...Retry) ([]byte, error) {
    var (
        err        error        // 读取错误
        size       int          // 读取大小
        data       []byte       // 缓冲对象
        remoteAddr *net.UDPAddr // 当前远程读取地址
    )
    if buffer > 0 {
        data = make([]byte, buffer)
    } else {
        data = make([]byte, defaultReadBufferSize)
    }
    for {
        size, remoteAddr, err = c.ReadFromUDP(data)
        if err == nil {
            c.remoteAddr = remoteAddr
        }
        if err != nil {
            // 连接关闭。
            if err == io.EOF {
                break
            }
            if len(retry) > 0 {
                // 即使重试也失败了。
                if retry[0].Count == 0 {
                    break
                }
                retry[0].Count--
                if retry[0].Interval == 0 {
                    retry[0].Interval = defaultRetryInterval
                }
                time.Sleep(retry[0].Interval)
                continue
            }
            err = errors.New("ReadFromUDP failed")
            break
        }
        break
    }
    return data[:size], err
}

// SendReceive 将数据写入连接并阻止读取响应。
func (c *Conn) SendReceive(data []byte, receive int, retry ...Retry) ([]byte, error) {
    if err := c.Send(data, retry...); err == nil {
        return c.Receive(receive, retry...)
    } else {
        return nil, err
    }
}

// ReceiveWithTimeout 从远程地址读取数据超时。
func (c *Conn) ReceiveWithTimeout(length int, timeout time.Duration, retry ...Retry) (data []byte, err error) {
    if err = c.SetReceiveDeadline(time.Now().Add(timeout)); err != nil {
        return nil, err
    }
    defer c.SetReceiveDeadline(time.Time{})
    data, err = c.Receive(length, retry...)
    return
}

// SendWithTimeout 将数据写入超时连接。
func (c *Conn) SendWithTimeout(data []byte, timeout time.Duration, retry ...Retry) (err error) {
    if err = c.SetSendDeadline(time.Now().Add(timeout)); err != nil {
        return err
    }
    defer c.SetSendDeadline(time.Time{})
    err = c.Send(data, retry...)
    return
}

// SendReceiveWithTimeout 将数据写入连接并读取超时响应。
func (c *Conn) SendReceiveWithTimeout(data []byte, receive int, timeout time.Duration, retry ...Retry) ([]byte, error) {
    if err := c.Send(data, retry...); err == nil {
        return c.ReceiveWithTimeout(receive, timeout, retry...)
    } else {
        return nil, err
    }
}

func (c *Conn) SetDeadline(t time.Time) (err error) {
    if err = c.UDPConn.SetDeadline(t); err == nil {
        c.receiveDeadline = t
        c.sendDeadline = t
    } else {
        err = fmt.Errorf(`SetDeadline for connection failed with "%s"`, t)
    }
    return err
}

func (c *Conn) SetReceiveDeadline(t time.Time) (err error) {
    if err = c.SetReadDeadline(t); err == nil {
        c.receiveDeadline = t
    } else {
        err = fmt.Errorf(`SetReadDeadline for connection failed with "%s"`, t)
    }
    return err
}

func (c *Conn) SetSendDeadline(t time.Time) (err error) {
    if err = c.SetWriteDeadline(t); err == nil {
        c.sendDeadline = t
    } else {
        err = fmt.Errorf(`SetWriteDeadline for connection failed with "%s"`, t)
    }
    return err
}

// SetReceiveBufferWait 从连接读取所有数据时设置缓冲区等待超时。
// 等待时间不能太长，否则可能会延迟从远程地址接收数据。
func (c *Conn) SetReceiveBufferWait(d time.Duration) {
    c.receiveBufferWait = d
}

// RemoteAddr 返回当前 UDP 连接的远程地址。
// 请注意，它不能使用 c.conn.RemoteAddr()，因为它是 nil。
func (c *Conn) RemoteAddr() net.Addr {
    return c.remoteAddr
}

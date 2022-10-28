package gtcp

import (
    "bufio"
    "bytes"
    "crypto/tls"
    "errors"
    "fmt"
    "io"
    "net"
    "time"
)

const (
    // 读取缓冲区的默认间隔。
    receiveAllWaitTimeout = time.Millisecond
)

// Conn TCP 连接对象。
type Conn struct {
    net.Conn                        // 底层 TCP 连接对象。
    reader            *bufio.Reader // 用于连接的缓冲区读取器。
    receiveDeadline   time.Time     // 读取超时点。
    sendDeadline      time.Time     // 写入超时点。
    receiveBufferWait time.Duration // 读取缓冲区的间隔持续时间。
}

// NewConn 创建并返回指定地址的新连接。
func NewConn(addr string, timeout ...time.Duration) (*Conn, error) {
    if conn, err := NewNetConn(addr, timeout...); err == nil {
        return NewConnByNetConn(conn), nil
    } else {
        return nil, err
    }
}

// NewConnTLS 创建并返回一个新的 TLS 连接
// 使用指定的地址和 TLS 配置。
func NewConnTLS(addr string, tlsConfig *tls.Config) (*Conn, error) {
    if conn, err := NewNetConnTLS(addr, tlsConfig); err == nil {
        return NewConnByNetConn(conn), nil
    } else {
        return nil, err
    }
}

// NewConnByNetConn 使用指定的 net.Conn 对象创建并返回 TCP 连接对象。
func NewConnByNetConn(conn net.Conn) *Conn {
    return &Conn{
        Conn:              conn,
        reader:            bufio.NewReader(conn),
        receiveDeadline:   time.Time{},
        sendDeadline:      time.Time{},
        receiveBufferWait: receiveAllWaitTimeout,
    }
}

// Send 将数据写入远程地址。
func (c *Conn) Send(data []byte, retry ...Retry) error {
    for {
        if _, err := c.Write(data); err != nil {
            // 连接关闭。
            if err == io.EOF {
                return err
            }
            // 即使重试后仍然失败。
            if len(retry) == 0 || retry[0].Count == 0 {
                err = errors.New(`write data failed`)
                return err
            }
            if len(retry) > 0 {
                retry[0].Count--
                if retry[0].Interval == 0 {
                    retry[0].Interval = defaultRetryInternal
                }
                time.Sleep(retry[0].Interval)
            }
        } else {
            return nil
        }
    }
}

// Receive 从连接中接收和返回数据。
//
// 注意，
// 1. 如果length = 0，表示从当前缓冲区接收数据并立即返回。
// 2. 如果length < 0，表示从connection接收所有数据，直到没有数据才返回
// 从连接。 如果您决定从缓冲区接收所有数据，开发人员应该注意自己解析的包。
// 3. 如果length > 0，这意味着它阻止从连接中读取数据，直到收到长度大小。 它是数据接收最常用的长度值。
func (c *Conn) Receive(length int, retry ...Retry) ([]byte, error) {
    var (
        err        error  // 读取错误。
        size       int    // 读取大小。
        index      int    // 接收大小。
        buffer     []byte // 缓冲对象。
        bufferWait bool   // 是否设置缓冲区读取超时。
    )
    if length > 0 {
        buffer = make([]byte, length)
    } else {
        buffer = make([]byte, defaultReadBufferSize)
    }

    for {
        if length < 0 && index > 0 {
            bufferWait = true
            if err = c.SetReadDeadline(time.Now().Add(c.receiveBufferWait)); err != nil {
                err = fmt.Errorf(`SetReadDeadline for connection failed`)
                return nil, err
            }
        }
        size, err = c.reader.Read(buffer[index:])
        if size > 0 {
            index += size
            if length > 0 {
                // 如果指定了 `length`，它将读取直到 `length` 大小。
                if index == length {
                    break
                }
            } else {
                if index >= defaultReadBufferSize {
                    // 如果超过缓冲区大小，它会自动增加其缓冲区大小。
                    buffer = append(buffer, make([]byte, defaultReadBufferSize)...)
                } else {
                    // 如果接收到的大小小于缓冲区大小，它会立即返回。
                    if !bufferWait {
                        break
                    }
                }
            }
        }
        if err != nil {
            // 连接关闭。
            if err == io.EOF {
                break
            }
            // 读取数据时重新设置超时。
            if bufferWait && isTimeout(err) {
                if err = c.SetReadDeadline(c.receiveDeadline); err != nil {
                    err = errors.New(`SetReadDeadline for connection failed`)
                    return nil, err
                }
                err = nil
                break
            }
            if len(retry) > 0 {
                // 即使重试也失败了。
                if retry[0].Count == 0 {
                    break
                }
                retry[0].Count--
                if retry[0].Interval == 0 {
                    retry[0].Interval = defaultRetryInternal
                }
                time.Sleep(retry[0].Interval)
                continue
            }
            break
        }
        // 只需从缓冲区读取一次。
        if length == 0 {
            break
        }
    }
    return buffer[:index], err
}

// ReceiveLine 从连接中读取数据，直到读取字符 '\n'。
// 请注意，返回的结果不包含最后一个字符 '\n'。
func (c *Conn) ReceiveLine(retry ...Retry) ([]byte, error) {
    var (
        err    error
        buffer []byte
        data   = make([]byte, 0)
    )
    for {
        buffer, err = c.Receive(1, retry...)
        if len(buffer) > 0 {
            if buffer[0] == '\n' {
                data = append(data, buffer[:len(buffer)-1]...)
                break
            } else {
                data = append(data, buffer...)
            }
        }
        if err != nil {
            break
        }
    }
    return data, err
}

// ReceiveTill 从连接中读取数据，直到读取字节`til`。
// 请注意，返回的结果包含最后一个字节`til`。
func (c *Conn) ReceiveTill(til []byte, retry ...Retry) ([]byte, error) {
    var (
        err    error
        buffer []byte
        data   = make([]byte, 0)
        length = len(til)
    )
    for {
        buffer, err = c.Receive(1, retry...)
        if len(buffer) > 0 {
            if length > 0 &&
                len(data) >= length-1 &&
                buffer[0] == til[length-1] &&
                bytes.EqualFold(data[len(data)-length+1:], til[:length-1]) {
                data = append(data, buffer...)
                break
            } else {
                data = append(data, buffer...)
            }
        }
        if err != nil {
            break
        }
    }
    return data, err
}

// ReceiveWithTimeout 从超时的连接中读取数据。
func (c *Conn) ReceiveWithTimeout(length int, timeout time.Duration, retry ...Retry) (data []byte, err error) {
    if err = c.SetReceiveDeadline(time.Now().Add(timeout)); err != nil {
        return nil, err
    }
    defer c.SetReceiveDeadline(time.Time{})
    data, err = c.Receive(length, retry...)
    return
}

// SendWithTimeout 将数据写入超时的连接。
func (c *Conn) SendWithTimeout(data []byte, timeout time.Duration, retry ...Retry) (err error) {
    if err = c.SetSendDeadline(time.Now().Add(timeout)); err != nil {
        return err
    }
    defer c.SetSendDeadline(time.Time{})
    err = c.Send(data, retry...)
    return
}

// SendReceive 将数据写入连接并阻止读取响应。
func (c *Conn) SendReceive(data []byte, length int, retry ...Retry) ([]byte, error) {
    if err := c.Send(data, retry...); err == nil {
        return c.Receive(length, retry...)
    } else {
        return nil, err
    }
}

// SendReceiveWithTimeout 将数据写入连接并读取超时响应。
func (c *Conn) SendReceiveWithTimeout(data []byte, length int, timeout time.Duration, retry ...Retry) ([]byte, error) {
    if err := c.Send(data, retry...); err == nil {
        return c.ReceiveWithTimeout(length, timeout, retry...)
    } else {
        return nil, err
    }
}

func (c *Conn) SetDeadline(t time.Time) (err error) {
    if err = c.Conn.SetDeadline(t); err == nil {
        c.receiveDeadline = t
        c.sendDeadline = t
    }
    if err != nil {
        err = fmt.Errorf(`SetDeadline for connection failed with "%s"`, t)
    }
    return err
}

func (c *Conn) SetReceiveDeadline(t time.Time) (err error) {
    if err = c.SetReadDeadline(t); err == nil {
        c.receiveDeadline = t
    }
    if err != nil {
        err = fmt.Errorf(`SetReadDeadline for connection failed with "%s"`, t)
    }
    return err
}

func (c *Conn) SetSendDeadline(t time.Time) (err error) {
    if err = c.SetWriteDeadline(t); err == nil {
        c.sendDeadline = t
    }
    if err != nil {
        err = fmt.Errorf(`SetWriteDeadline for connection failed with "%s"`, t)
    }
    return err
}

// SetReceiveBufferWait 从连接读取所有数据时设置缓冲区等待超时。
// 等待时间不能太长，否则可能会延迟从远程地址接收数据。
func (c *Conn) SetReceiveBufferWait(bufferWaitDuration time.Duration) {
    c.receiveBufferWait = bufferWaitDuration
}

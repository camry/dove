package gtcp

import (
    "encoding/binary"
    "fmt"
    "time"
)

const (
    pkgHeaderSizeDefault = 2 // 包协议的标头大小。
    pkgHeaderSizeMax     = 4 // 包协议的最大标头大小。
)

// PkgOption 是协议的封装选项。
type PkgOption struct {
    // HeaderSize 用于标记下一次数据接收的数据长度。
    // 默认为 2 字节，最大 4 字节，表示最大数据长度为 65535 到 4294967295 字节。
    HeaderSize int

    // MaxDataSize 是用于数据长度验证的数据字段大小（以字节为单位）。
    // 如果不手动设置，会自动与HeaderSize对应设置。
    MaxDataSize int

    // Retry 操作失败时的策略。
    Retry Retry
}

// SendPkg 使用包协议发送数据。
//
// 包协议：DataLength(24bit)|DataField(variant)。
//
// 注意，
// 1. DataLength是DataField的长度，不包含header的大小。
// 2. 包的整数字节使用 BigEndian 顺序编码。
func (c *Conn) SendPkg(data []byte, option ...PkgOption) error {
    pkgOption, err := getPkgOption(option...)
    if err != nil {
        return err
    }
    length := len(data)
    if length > pkgOption.MaxDataSize {
        return fmt.Errorf(
            `data too long, data size %d exceeds allowed max data size %d`,
            length, pkgOption.MaxDataSize,
        )
    }
    offset := pkgHeaderSizeMax - pkgOption.HeaderSize
    buffer := make([]byte, pkgHeaderSizeMax+len(data))
    binary.BigEndian.PutUint32(buffer[0:], uint32(length))
    copy(buffer[pkgHeaderSizeMax:], data)
    if pkgOption.Retry.Count > 0 {
        return c.Send(buffer[offset:], pkgOption.Retry)
    }
    return c.Send(buffer[offset:])
}

// SendPkgWithTimeout 使用包协议将数据写入超时连接。
func (c *Conn) SendPkgWithTimeout(data []byte, timeout time.Duration, option ...PkgOption) (err error) {
    if err := c.SetSendDeadline(time.Now().Add(timeout)); err != nil {
        return err
    }
    defer c.SetSendDeadline(time.Time{})
    err = c.SendPkg(data, option...)
    return
}

// SendReceivePkg 使用包协议将数据写入连接并阻止读取响应。
func (c *Conn) SendReceivePkg(data []byte, option ...PkgOption) ([]byte, error) {
    if err := c.SendPkg(data, option...); err == nil {
        return c.ReceivePkg(option...)
    } else {
        return nil, err
    }
}

// SendReceivePkgWithTimeout 使用包协议将数据写入连接并读取超时响应。
func (c *Conn) SendReceivePkgWithTimeout(data []byte, timeout time.Duration, option ...PkgOption) ([]byte, error) {
    if err := c.SendPkg(data, option...); err == nil {
        return c.ReceivePkgWithTimeout(timeout, option...)
    } else {
        return nil, err
    }
}

// ReceivePkg 使用包协议从连接接收数据。
func (c *Conn) ReceivePkg(option ...PkgOption) (result []byte, err error) {
    var (
        buffer []byte
        length int
    )
    pkgOption, err := getPkgOption(option...)
    if err != nil {
        return nil, err
    }
    // 头字段。
    buffer, err = c.Receive(pkgOption.HeaderSize, pkgOption.Retry)
    if err != nil {
        return nil, err
    }
    switch pkgOption.HeaderSize {
    case 1:
        // 如果标头大小小于 4 个字节 (uint32)，则填充为零。
        length = int(binary.BigEndian.Uint32([]byte{0, 0, 0, buffer[0]}))
    case 2:
        length = int(binary.BigEndian.Uint32([]byte{0, 0, buffer[0], buffer[1]}))
    case 3:
        length = int(binary.BigEndian.Uint32([]byte{0, buffer[0], buffer[1], buffer[2]}))
    default:
        length = int(binary.BigEndian.Uint32([]byte{buffer[0], buffer[1], buffer[2], buffer[3]}))
    }
    // 这里验证包的大小。
    // 如果验证失败，它会清除缓冲区并立即返回错误。
    if length < 0 || length > pkgOption.MaxDataSize {
        return nil, fmt.Errorf(`invalid package size %d`, length)
    }
    // 空包。
    if length == 0 {
        return nil, nil
    }
    // 数据字段。
    return c.Receive(length, pkgOption.Retry)
}

// ReceivePkgWithTimeout 使用包协议从超时连接中读取数据。
func (c *Conn) ReceivePkgWithTimeout(timeout time.Duration, option ...PkgOption) (data []byte, err error) {
    if err := c.SetReceiveDeadline(time.Now().Add(timeout)); err != nil {
        return nil, err
    }
    defer c.SetReceiveDeadline(time.Time{})
    data, err = c.ReceivePkg(option...)
    return
}

// getPkgOption 包装并返回 PkgOption。
// 如果没有给出选项，它返回一个具有默认值的新选项。
func getPkgOption(option ...PkgOption) (*PkgOption, error) {
    pkgOption := PkgOption{}
    if len(option) > 0 {
        pkgOption = option[0]
    }
    if pkgOption.HeaderSize == 0 {
        pkgOption.HeaderSize = pkgHeaderSizeDefault
    }
    if pkgOption.HeaderSize > pkgHeaderSizeMax {
        return nil, fmt.Errorf(
            `package header size %d definition exceeds max header size %d`,
            pkgOption.HeaderSize, pkgHeaderSizeMax,
        )
    }
    if pkgOption.MaxDataSize == 0 {
        switch pkgOption.HeaderSize {
        case 1:
            pkgOption.MaxDataSize = 0xFF
        case 2:
            pkgOption.MaxDataSize = 0xFFFF
        case 3:
            pkgOption.MaxDataSize = 0xFFFFFF
        case 4:
            // math.MaxInt32 不是 math.MaxUint32
            pkgOption.MaxDataSize = 0x7FFFFFFF
        }
    }
    if pkgOption.MaxDataSize > 0x7FFFFFFF {
        return nil, fmt.Errorf(
            `package data size %d definition exceeds allowed max data size %d`,
            pkgOption.MaxDataSize, 0x7FFFFFFF,
        )
    }
    return &pkgOption, nil
}

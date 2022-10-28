package gtcp

import "time"

// SendPkg 将包含 `data` 的包发送到 `address` 并关闭连接。
// 可选参数 `option` 指定发送的包选项。
func SendPkg(address string, data []byte, option ...PkgOption) error {
    conn, err := NewConn(address)
    if err != nil {
        return err
    }
    defer conn.Close()
    return conn.SendPkg(data, option...)
}

// SendReceivePkg 将包含 `data` 的包发送到 `address`，接收响应并关闭连接。
// 可选参数 `option` 指定发送的包选项。
func SendReceivePkg(address string, data []byte, option ...PkgOption) ([]byte, error) {
    conn, err := NewConn(address)
    if err != nil {
        return nil, err
    }
    defer conn.Close()
    return conn.SendReceivePkg(data, option...)
}

// SendPkgWithTimeout 将包含 `data` 的包发送到具有超时限制的 `address` 并关闭连接。
// 可选参数 `option` 指定发送的包选项。
func SendPkgWithTimeout(address string, data []byte, timeout time.Duration, option ...PkgOption) error {
    conn, err := NewConn(address)
    if err != nil {
        return err
    }
    defer conn.Close()
    return conn.SendPkgWithTimeout(data, timeout, option...)
}

// SendReceivePkgWithTimeout 将包含 `data` 的包发送到 `address` ，接收具有超时限制的响应并关闭连接。
// 可选参数 `option` 指定发送的包选项。
func SendReceivePkgWithTimeout(address string, data []byte, timeout time.Duration, option ...PkgOption) ([]byte, error) {
    conn, err := NewConn(address)
    if err != nil {
        return nil, err
    }
    defer conn.Close()
    return conn.SendReceivePkgWithTimeout(data, timeout, option...)
}

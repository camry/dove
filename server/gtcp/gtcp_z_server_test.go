package gtcp_test

import (
    "context"
    "fmt"
    "reflect"
    "testing"
    "time"

    "github.com/camry/dove/server/gtcp"
)

var (
    simpleTimeout = time.Millisecond * 100
    sendData      = []byte("hello")
)

func TestNewServer(t *testing.T) {
    var (
        port, _ = gtcp.GetFreePort()
        ctx     = context.Background()
        addr    = fmt.Sprintf("127.0.0.1:%d", port)
    )
    s := gtcp.NewServer(gtcp.Address(addr), gtcp.Handler(func(conn *gtcp.Conn) {
        defer conn.Close()
        for {
            data, err := conn.Receive(-1)
            if err != nil {
                break
            }
            conn.Send(data)
        }
    }))
    go s.Start(ctx)
    time.Sleep(simpleTimeout)
    receive, err := gtcp.SendReceive(addr, sendData, -1)
    if err != nil {
        t.Error(err)
    }
    if !reflect.DeepEqual(receive, sendData) {
        t.Fatalf("%s is not equal to v:%s", string(receive), string(sendData))
    }
}

func TestNewPkgServer(t *testing.T) {
    var (
        port, _ = gtcp.GetFreePort()
        ctx     = context.Background()
        addr    = fmt.Sprintf("127.0.0.1:%d", port)
    )
    s := gtcp.NewServer(gtcp.Address(addr), gtcp.Handler(func(conn *gtcp.Conn) {
        defer conn.Close()
        for {
            data, err := conn.ReceivePkg()
            if err != nil {
                break
            }
            conn.SendPkg(data)
        }
    }))
    go s.Start(ctx)
    time.Sleep(simpleTimeout)
    receive, err := gtcp.SendReceivePkg(addr, sendData)
    if err != nil {
        t.Error(err)
    }
    if !reflect.DeepEqual(receive, sendData) {
        t.Fatalf("%s is not equal to v:%s", string(receive), string(sendData))
    }
}

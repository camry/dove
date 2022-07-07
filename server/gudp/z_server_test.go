package gudp_test

import (
    "context"
    "fmt"
    "github.com/camry/dove/log"
    "github.com/camry/dove/server/gudp"
    "reflect"
    "strconv"
    "testing"
    "time"
)

var (
    simpleTimeout = time.Millisecond * 100
    sendData      = []byte("hello")
)

func TestNewConn(t *testing.T) {
    var (
        port, _ = gudp.GetFreePort()
        ctx     = context.Background()
    )

    s := gudp.NewServer(gudp.Address(fmt.Sprintf("127.0.0.1:%d", port)), gudp.Handler(func(conn *gudp.Conn) {
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

    conn, err := gudp.NewConn(fmt.Sprintf("127.0.0.1:%d", port), fmt.Sprintf("127.0.0.1:%d", port+1))
    if err != nil {
        t.Error(err)
    }
    err1 := conn.SetDeadline(time.Now().Add(time.Second))
    if err != nil {
        t.Error(err1)
    }
    err2 := conn.Send(sendData)
    if err != nil {
        t.Error(err2)
    }
    err3 := conn.Close()
    if err != nil {
        t.Error(err3)
    }
}

func TestNewServer(t *testing.T) {
    var (
        ctx = context.Background()
    )
    p, _ := gudp.GetFreePort()
    s := gudp.NewServer(gudp.Address(fmt.Sprintf("127.0.0.1:%d", p)), gudp.Handler(func(conn *gudp.Conn) {
        logger := log.NewHelper(log.GetLogger())
        defer conn.Close()
        for {
            data, err := conn.Receive(-1)
            if len(data) > 0 {
                if err := conn.Send(append([]byte("> "), data...)); err != nil {
                    logger.Error(ctx, err)
                }
            }
            if err != nil {
                break
            }
        }
    }))
    go s.Start(ctx)
    defer s.Stop(ctx)
    time.Sleep(100 * time.Millisecond)

    // gudp.Conn.Send
    for i := 0; i < 100; i++ {
        conn, err1 := gudp.NewConn(fmt.Sprintf("127.0.0.1:%d", p))
        if err1 != nil {
            t.Error(err1)
        }
        err2 := conn.Send([]byte(strconv.Itoa(i)))
        if err2 != nil {
            t.Error(err2)
        }
        result, err3 := conn.Receive(-1)
        if err3 != nil {
            t.Error(err3)
        }
        if reflect.DeepEqual(conn.RemoteAddr(), nil) {
            t.Fatalf("conn.RemoteAddr():%v is equal to v:%v", conn.RemoteAddr(), nil)
        }
        if !reflect.DeepEqual(string(result), fmt.Sprintf(`> %d`, i)) {
            t.Fatalf("%s is not equal to v:%s", string(result), fmt.Sprintf(`> %d`, i))
        }
        conn.Close()
    }
}

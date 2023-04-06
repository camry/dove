package dove

import (
    "context"
    "reflect"
    "testing"
    "time"

    "github.com/camry/dove/server/gcron"
    "github.com/camry/dove/server/ghttp"
    "github.com/camry/dove/server/grpc"
    "github.com/camry/dove/server/gtcp"
    "github.com/camry/dove/server/gudp"
    ggtcp "github.com/camry/g/gnet/gtcp"
    ggudp "github.com/camry/g/gnet/gudp"
)

func TestNew(t *testing.T) {
    hs := ghttp.NewServer()
    gs := grpc.NewServer()
    gc := gcron.NewServer()
    tcp := gtcp.NewServer(gtcp.Handler(func(conn *ggtcp.Conn) {
    }))
    udp := gudp.NewServer(gudp.Handler(func(conn *ggudp.Conn) {
    }))
    app := New(
        Name("dove"),
        Version(Release),
        Server(hs, gs, gc, tcp, udp),
        BeforeStart(func(_ context.Context) error {
            t.Log("BeforeStart...")
            return nil
        }),
        BeforeStop(func(_ context.Context) error {
            t.Log("BeforeStop...")
            return nil
        }),
        AfterStart(func(_ context.Context) error {
            t.Log("AfterStart...")
            return nil
        }),
        AfterStop(func(_ context.Context) error {
            t.Log("AfterStop...")
            return nil
        }),
    )
    time.AfterFunc(time.Second, func() {
        _ = app.Stop()
    })
    if err := app.Run(); err != nil {
        t.Fatal(err)
    }
}

func TestApp_ID(t *testing.T) {
    v := "123"
    o := New(ID(v))
    if !reflect.DeepEqual(v, o.ID()) {
        t.Fatalf("o.ID():%s is not equal to v:%s", o.ID(), v)
    }
}

func TestApp_Name(t *testing.T) {
    v := "123"
    o := New(Name(v))
    if !reflect.DeepEqual(v, o.Name()) {
        t.Fatalf("o.Name():%s is not equal to v:%s", o.Name(), v)
    }
}

func TestApp_Version(t *testing.T) {
    v := "123"
    o := New(Version(v))
    if !reflect.DeepEqual(v, o.Version()) {
        t.Fatalf("o.Version():%s is not equal to v:%s", o.Version(), v)
    }
}

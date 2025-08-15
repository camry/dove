# dove

应用程序生命周期服务管理框架。

## 安装

```bash
go get -u github.com/camry/dove/v2
```

## 使用

```go
package main

import (
    "context"
    "log"

    "github.com/camry/dove/v2"
    "github.com/camry/dove/v2/server/gcron"
    "github.com/camry/dove/v2/server/ghttp"
    "github.com/camry/dove/v2/server/grpc"
    "github.com/camry/dove/v2/server/gtcp"
    "github.com/camry/dove/v2/server/gudp"
    ggtcp "github.com/camry/g/v2/gnet/gtcp"
    ggudp "github.com/camry/g/v2/gnet/gudp"
)

func main() {
    hs := ghttp.NewServer()
    gs := grpc.NewServer()
    gc := gcron.NewServer()
    tcp := gtcp.NewServer(gtcp.Handler(func(conn *ggtcp.Conn) {
    }))
    udp := gudp.NewServer(gudp.Handler(func(conn *ggudp.ServerConn) {
    }))
    app := dove.New(
        dove.Name("dove"),
        dove.Version(dove.Release),
        dove.Server(hs, gs, gc, tcp, udp),
        dove.BeforeStart(func(_ context.Context) error {
            log.Println("BeforeStart...")
            return nil
        }),
        dove.BeforeStop(func(_ context.Context) error {
            log.Println("BeforeStop...")
            return nil
        }),
        dove.AfterStart(func(_ context.Context) error {
            log.Println("AfterStart...")
            return nil
        }),
        dove.AfterStop(func(_ context.Context) error {
            log.Println("AfterStop...")
            return nil
        }),
    )
    if err := app.Run(); err != nil {
        log.Fatal(err)
    }
}
```

## License

Dove 框架开源许可证 [MIT LICENSE](https://github.com/camry/g/v2/blob/main/LICENSE)

## 致谢

以下项目对 Dove 框架的设计产生了特别的影响。

- [go-kratos/kratos](https://github.com/go-kratos/kratos) Kratos is a microservice-oriented governance framework
  implemented by golang.

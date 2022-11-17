# dove

应用程序生命周期服务管理框架。

## 安装

```bash
go get -u github.com/camry/dove
```

## 使用

```go
package main

import (
	"log"

    "github.com/camry/dove"
    "github.com/camry/dove/server/gcron"
    "github.com/camry/dove/server/ghttp"
    "github.com/camry/dove/server/grpc"
    "github.com/camry/dove/server/gtcp"
    "github.com/camry/dove/server/gudp"
    ggtcp "github.com/camry/g/gnet/gtcp"
    ggudp "github.com/camry/g/gnet/gudp"
)

func main() {
    hs := ghttp.NewServer()
    gs := grpc.NewServer()
    gc := gcron.NewServer()
    tcp := gtcp.NewServer(gtcp.Handler(func(conn *ggtcp.Conn) {
    }))
    udp := gudp.NewServer(gudp.Handler(func(conn *ggudp.Conn) {
    }))
    app := dove.New(
        dove.Name("dove"),
        dove.Version(dove.Release),
        dove.Server(hs, gs, gc, tcp, udp),
    )
    if err := app.Run(); err != nil {
        log.Fatal(err)
    }
}
```

## License

Dove 框架开源许可证 [MIT LICENSE](https://github.com/camry/g/blob/main/LICENSE)

## 致谢

- [go-kratos/kratos](https://github.com/go-kratos/kratos) Kratos 一套轻量级 Go 微服务框架，包含大量微服务相关功能及工具。

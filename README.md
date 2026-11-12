# hqpcb-go-sdk

HQ PCB（PCB_V5）Go SDK：Guest（TinyGo WASI / `PCB_HOST_DIR` 文件桥）与本机 Arrow Flight IPC 客户端。

## Install

```bash
go get github.com/Huaqiu-Electronics/hqpcb-go-sdk@latest
```

## Import

```go
import "github.com/Huaqiu-Electronics/hqpcb-go-sdk"

func main() {
    pcb.Log("hello")
    out, err := pcb.WorkspaceList()
    _ = out
    _ = err
}
```

包名仍为 `pcb`（与历史 `hq.pcb/sdk/pcb` 一致）；模块路径已迁到本仓库。

## Guest 脚本（PCB_V5 宿主）

桌面应用通过 `script.go.*` 编译运行时，会把 `go.mod` 的 `require`/`replace` 指到捆绑的 `go-sdk/pcb`（本仓库内容的拷贝）。用户源码请使用：

```go
import "github.com/Huaqiu-Electronics/hqpcb-go-sdk"
```

## 文档

见 PCB_V5 仓库：`docs/07-插件与扩展/01-宿主与协议/GoScriptPlatform.md`、`sdk/README.md`。

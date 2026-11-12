// ipc_smoke：对本机 pcb_IPC_skel_serve 跑 Flight hello/ping/caps（P6d）。
//
// 用法：
//
//	ipc_smoke                 # 默认逻辑名 hqpcb-svc
//	ipc_smoke hqpcb-svc       # 显式逻辑名（不要传 socket 路径 / URI）
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/Huaqiu-Electronics/hqpcb-go-sdk"
)

func main() {
	pipe := ""
	if len(os.Args) >= 2 {
		arg := strings.TrimSpace(os.Args[1])
		if strings.Contains(arg, "/") || strings.Contains(arg, "://") || strings.HasSuffix(arg, ".sock") {
			fmt.Fprintln(os.Stderr, "usage: ipc_smoke [logical-pipe-name]")
			fmt.Fprintln(os.Stderr, "  do not pass Flight URI / socket path; use hqpcb-svc")
			os.Exit(2)
		}
		pipe = arg
	}
	c := pcb.NewFlightIPC()
	if err := c.Connect(pipe, false); err != nil {
		fmt.Fprintln(os.Stderr, "connect:", err)
		os.Exit(1)
	}
	defer c.Disconnect()
	echo, err := c.Ping("p6d-go")
	if err != nil {
		fmt.Fprintln(os.Stderr, "ping:", err)
		os.Exit(1)
	}
	caps, err := c.Caps()
	if err != nil {
		fmt.Fprintln(os.Stderr, "caps:", err)
		os.Exit(1)
	}
	cat, err := c.MethodsList()
	if err != nil {
		fmt.Fprintln(os.Stderr, "methods.list:", err)
		os.Exit(1)
	}
	if !strings.Contains(cat, "objects.list") || !strings.Contains(cat, "transport=doget") {
		fmt.Fprintln(os.Stderr, "methods.list missing P6c doget entries")
		os.Exit(1)
	}
	fmt.Printf("OK pipe=%s role=%s echo=%s workspace=%v design=%v ver=%d.%d\n",
		c.Pipe(), c.PeerRole(), echo, caps.Workspace, caps.Design, caps.VerMajor, caps.VerMinor)
}

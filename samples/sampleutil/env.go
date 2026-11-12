// Package sampleutil 为 Go Flight samples 共用参数解析（headless / gui）。
package sampleutil

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Endpoint 解析结果：逻辑管道名 + 是否 GUI。
type Endpoint struct {
	Pipe string // 空则由 Connect 按 Gui 选默认名 / 环境变量
	Gui  bool
	Rest []string
}

func usageEndpoint() string {
	return "端点（默认 headless）：\n" +
		"  headless | service | svc | hqpcb-svc | --headless\n" +
		"  gui | hqpcb-gui | --gui\n" +
		"  --endpoint=headless|gui\n" +
		"环境变量：PCB_IPC_ENDPOINT=headless|gui；PCB_IPC_SERVICE_PIPE / PCB_IPC_GUI_PIPE\n"
}

// ResolvePipe 解析端点与剩余参数（兼容旧名）。
func ResolvePipe(argv []string) (pipe string, gui bool, rest []string) {
	ep := ResolveEndpoint(argv)
	return ep.Pipe, ep.Gui, ep.Rest
}

// ResolveEndpoint 解析 headless/gui 端点。
func ResolveEndpoint(argv []string) Endpoint {
	ep := Endpoint{Rest: nil}
	if env := strings.ToLower(strings.TrimSpace(os.Getenv("PCB_IPC_ENDPOINT"))); env != "" {
		switch env {
		case "gui", "hqpcb-gui":
			ep.Gui = true
		case "headless", "service", "svc", "hqpcb-svc":
			ep.Gui = false
		}
	}
	args := append([]string(nil), argv[1:]...)
	for _, a := range args {
		low := strings.ToLower(a)
		if strings.HasPrefix(a, "grpc") || strings.Contains(a, "://") || strings.HasSuffix(a, ".sock") {
			fmt.Fprintln(os.Stderr, "error: 不要传 Flight URI / socket。\n"+usageEndpoint())
			os.Exit(2)
		}
		if strings.HasPrefix(low, "--endpoint=") {
			val := strings.TrimSpace(strings.SplitN(low, "=", 2)[1])
			switch val {
			case "gui", "hqpcb-gui":
				ep.Gui, ep.Pipe = true, ""
			case "headless", "service", "svc", "hqpcb-svc":
				ep.Gui, ep.Pipe = false, ""
			default:
				fmt.Fprintf(os.Stderr, "error: unknown --endpoint=%s\n%s", val, usageEndpoint())
				os.Exit(2)
			}
			continue
		}
		switch low {
		case "gui", "--gui":
			ep.Gui, ep.Pipe = true, ""
			continue
		case "headless", "service", "svc", "--headless", "--service":
			ep.Gui, ep.Pipe = false, ""
			continue
		case "hqpcb-gui":
			ep.Gui, ep.Pipe = true, a
			continue
		case "hqpcb-svc":
			ep.Gui, ep.Pipe = false, a
			continue
		}
		ep.Rest = append(ep.Rest, a)
	}
	return ep
}

// DefaultBoardPath 优先 python/samples 旁 complex_hierarchy，再仓库 samples/。
func DefaultBoardPath() string {
	_, file, _, ok := runtime.Caller(0)
	root := "."
	if ok {
		root = filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", "..", ".."))
	}
	cands := []string{
		filepath.Join(root, "sdk", "python", "samples", "complex_hierarchy.hqpcb"),
		filepath.Join(root, "samples", "complex_hierarchy.hqpcb"),
	}
	for _, c := range cands {
		if st, e := os.Stat(c); e == nil && !st.IsDir() {
			abs, _ := filepath.Abs(c)
			return abs
		}
	}
	return cands[0]
}

// RequireWsDesign 在 board/ws 样例入口检查对端 caps；gui 端点无 workspace。
func RequireWsDesign(pipe, role string, workspace, design bool) {
	if workspace && design {
		return
	}
	fmt.Fprintf(os.Stderr,
		"peer lacks workspace/design (pipe=%s role=%s)\n"+
			"请连已声明 workspace+design 的端点（hqpcb-svc 或已升级的 hqpcb-gui）。\n",
		pipe, role)
	os.Exit(1)
}

// BoardPath 显式路径或默认板。
func BoardPath(arg string) (string, error) {
	if strings.TrimSpace(arg) == "" {
		p := DefaultBoardPath()
		if st, err := os.Stat(p); err != nil || st.IsDir() {
			return "", fmt.Errorf("board not found: %s", p)
		}
		return p, nil
	}
	abs, err := filepath.Abs(arg)
	if err != nil {
		return "", err
	}
	if st, err := os.Stat(abs); err != nil || st.IsDir() {
		return "", fmt.Errorf("board not found: %s", abs)
	}
	return abs, nil
}

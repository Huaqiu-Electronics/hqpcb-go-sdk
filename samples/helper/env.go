// Package helper：Go Flight samples 共用（板路径等）；端点解析用 hqeda.ResolveEndpoint。
package helper

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	hqeda "github.com/Huaqiu-Electronics/hqpcb-go-sdk"
)

// ResolveEndpoint 委托 SDK；失败时打印并 os.Exit(2)。
func ResolveEndpoint(argv []string) hqeda.Endpoint {
	ep, err := hqeda.ResolveEndpoint(argv)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(2)
	}
	return ep
}

// ResolvePipe 兼容旧样例签名。
func ResolvePipe(argv []string) (pipe string, gui bool, rest []string) {
	ep := ResolveEndpoint(argv)
	return ep.Pipe, ep.Gui, ep.Rest
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

// CheckEnv 在 board/ws 样例入口检查对端 workspace+design 能力。
func CheckEnv(pipe, role string, workspace, design bool) {
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

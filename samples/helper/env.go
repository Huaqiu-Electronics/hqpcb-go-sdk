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

// BoardPresets 示例名 → Samples/Boards/hqpcbx 文件名。
var BoardPresets = map[string]string{
	"tracks":      "01_tracks_vias.hqpcbx",
	"arc_tracks":  "05_arc_tracks_vias.hqpcbx",
	"components":  "02_components_pads.hqpcbx",
	"zones":       "03_zones_graphics.hqpcbx",
	"all_basics":  "04_all_basics.hqpcbx",
	"keepout":     "06_keepout_netclass.hqpcbx",
	"custom_pads": "07_custom_pads_curves.hqpcbx",
	"blind":       "08_blind_bottom.hqpcbx",
	"slots":       "09_slots_thermals.hqpcbx",
	"outline":     "10_outline_arcs.hqpcbx",
	"diffpair":    "11_diffpair_route.hqpcbx",
}

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

func repoRoot() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "."
	}
	// helper -> samples -> hqeda -> go -> sdk -> repo
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", "..", ".."))
}

// BoardsDir 返回 Samples/Boards/hqpcbx。
func BoardsDir() string {
	return filepath.Join(repoRoot(), "Samples", "Boards", "hqpcbx")
}

// DefaultBoardPath 默认综合小板（文本工程）。
func DefaultBoardPath() string {
	p := filepath.Join(BoardsDir(), BoardPresets["all_basics"])
	if abs, err := filepath.Abs(p); err == nil {
		return abs
	}
	return p
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

// BoardPath 显式路径或默认板（可传预设名如 tracks / all_basics）。
func BoardPath(arg string) (string, error) {
	if strings.TrimSpace(arg) == "" {
		p := DefaultBoardPath()
		if st, err := os.Stat(p); err != nil || st.IsDir() {
			return "", fmt.Errorf("board not found: %s (run Samples/Boards/convert_to_hqpcb.sh --hqpcbx-only)", p)
		}
		return p, nil
	}
	if name, ok := BoardPresets[arg]; ok {
		p := filepath.Join(BoardsDir(), name)
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p, nil
		}
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

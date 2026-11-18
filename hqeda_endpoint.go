package hqeda

import (
	"fmt"
	"os"
	"strings"
)

// Endpoint 为 CLI / 环境变量解析出的 Flight 逻辑端点。
//
// Pipe 为空时由 Connect(pipe, gui) 按角色选默认名（hqpcb-svc / hqpcb-gui）
// 或 PCB_IPC_SERVICE_PIPE / PCB_IPC_GUI_PIPE。
type Endpoint struct {
	Pipe string   // 显式逻辑名；可空
	Gui  bool     // true → GUI Interactive
	Rest []string // 未消费的位置参数（如板路径）
}

// ResolveEndpoint 解析 argv（通常为 os.Args）中的 headless/gui 端点。
//
// 支持：headless|service|svc|--headless、gui|--gui、hqpcb-svc|hqpcb-gui、
// --endpoint=…，以及环境变量 PCB_IPC_ENDPOINT。
// 拒绝 Flight URI / socket 路径。
func ResolveEndpoint(argv []string) (Endpoint, error) {
	ep := Endpoint{}
	if env := strings.ToLower(strings.TrimSpace(os.Getenv("PCB_IPC_ENDPOINT"))); env != "" {
		switch env {
		case "gui", "hqpcb-gui":
			ep.Gui = true
		case "headless", "service", "svc", "hqpcb-svc":
			ep.Gui = false
		}
	}
	args := argv
	if len(args) > 0 {
		args = args[1:]
	}
	for _, a := range args {
		low := strings.ToLower(a)
		if strings.HasPrefix(a, "grpc") || strings.Contains(a, "://") || strings.HasSuffix(a, ".sock") {
			return Endpoint{}, fmt.Errorf("不要传 Flight URI / socket；请用逻辑名 headless|gui|hqpcb-svc|hqpcb-gui")
		}
		if strings.HasPrefix(low, "--endpoint=") {
			val := strings.TrimSpace(strings.SplitN(low, "=", 2)[1])
			switch val {
			case "gui", "hqpcb-gui":
				ep.Gui, ep.Pipe = true, ""
			case "headless", "service", "svc", "hqpcb-svc":
				ep.Gui, ep.Pipe = false, ""
			default:
				return Endpoint{}, fmt.Errorf("unknown --endpoint=%s", val)
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
	return ep, nil
}

// ResolveEndpointPipe 等价于 ResolveEndpoint，拆成 pipe/gui/rest（便于样例）。
func ResolveEndpointPipe(argv []string) (pipe string, gui bool, rest []string, err error) {
	ep, err := ResolveEndpoint(argv)
	if err != nil {
		return "", false, nil, err
	}
	return ep.Pipe, ep.Gui, ep.Rest, nil
}

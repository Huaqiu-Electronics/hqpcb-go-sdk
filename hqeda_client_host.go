//go:build wasip1

// 脚本 / WASI 模式 ClientFactory：连接走 PCB_HOST_DIR；后续方法名对齐 Flight ClientFactory。
package hqeda

import (
	"fmt"
	"strconv"
	"strings"
)

// FlightCaps 为 caps 行协议解析结果（wasip1 侧本地定义；本机见 hqeda_ipc.go）。
type FlightCaps struct {
	Interactive bool
	Workspace   bool
	Design      bool
	VerMajor    int
	VerMinor    int
	Raw         map[string]string
}

// ParseCapsText 解析 hello/caps 行协议。
func ParseCapsText(text string) FlightCaps {
	raw := map[string]string{}
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || !strings.Contains(line, "=") {
			continue
		}
		k, v, _ := strings.Cut(line, "=")
		raw[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	flag := func(name string) bool {
		v := strings.ToLower(raw[name])
		return v == "1" || v == "true" || v == "yes"
	}
	maj, _ := strconv.Atoi(raw["ver_major"])
	if maj == 0 {
		maj = 1
	}
	min, _ := strconv.Atoi(raw["ver_minor"])
	return FlightCaps{
		Interactive: flag("interactive"),
		Workspace:   flag("workspace"),
		Design:      flag("design"),
		VerMajor:    maj,
		VerMinor:    min,
		Raw:         raw,
	}
}

// ClientFactory 是脚本模式产品入口（外观对齐本机 Flight ClientFactory）。
//
// 典型用法：
//
//	cli := hqeda.NewClientFactory()
//	_ = cli.Connect("", true) // 校验 PCB_HOST_DIR
//	_, _ = cli.WsOpen("/path/board.hqpcb")
type ClientFactory struct {
	connected bool
	pipe      string
	peerRole  string
}

// NewClientFactory 创建未连接的脚本模式客户端工厂。
func NewClientFactory() *ClientFactory {
	return &ClientFactory{pipe: "host-bridge", peerRole: "gui"}
}

func (c *ClientFactory) ensure() error {
	if c == nil || !c.connected {
		return fmt.Errorf("not connected (call Connect / ConnectGUI first)")
	}
	return nil
}

// Connect 绑定宿主文件桥（不建立 Flight socket）。
func (c *ClientFactory) Connect(pipe string, gui bool) error {
	if c == nil {
		return fmt.Errorf("nil ClientFactory")
	}
	if _, err := hostDir(); err != nil {
		return err
	}
	c.connected = true
	c.pipe = "host-bridge"
	if gui {
		c.peerRole = "gui"
	} else {
		c.peerRole = "service"
	}
	return nil
}

// ConnectService 对齐 Flight：脚本模式下仍只绑定宿主桥。
func (c *ClientFactory) ConnectService(pipe string) error {
	return c.Connect(pipe, false)
}

// ConnectGUI 对齐 Flight：脚本模式下仍只绑定宿主桥。
func (c *ClientFactory) ConnectGUI(pipe string) error {
	return c.Connect(pipe, true)
}

// Disconnect 清除已连接标记。
func (c *ClientFactory) Disconnect() {
	if c != nil {
		c.connected = false
	}
}

// Pipe 返回逻辑端点名（脚本模式恒为 host-bridge）。
func (c *ClientFactory) Pipe() string {
	if c == nil {
		return ""
	}
	return c.pipe
}

// DoAction 经宿主文件桥转发 Flight 同名动作。
func (c *ClientFactory) DoAction(method, body string) (string, error) {
	if err := c.ensure(); err != nil {
		return "", err
	}
	return call("doAction", map[string]interface{}{
		"method": method,
		"body":   body,
	})
}

// Ping 探活回显。
func (c *ClientFactory) Ping(token string) (string, error) {
	if token == "" {
		token = "go"
	}
	return c.DoAction("ping", token)
}

// Caps 查询能力行协议（解析为 FlightCaps）。
func (c *ClientFactory) Caps() (FlightCaps, error) {
	text, err := c.DoAction("caps", "")
	if err != nil {
		return FlightCaps{}, err
	}
	return ParseCapsText(text), nil
}

// MethodsList 方法目录。
func (c *ClientFactory) MethodsList() (string, error) {
	return c.DoAction("methods.list", "")
}

// ConfigGet 读配置。
func (c *ClientFactory) ConfigGet(key string) (string, error) {
	return c.DoAction("config.get", key)
}

// ConfigSet 写配置。
func (c *ClientFactory) ConfigSet(key, value string) (string, error) {
	return c.DoAction("config.set", key+"="+value)
}

// WsOpen 打开工程（宿主映射为 project.open）。
func (c *ClientFactory) WsOpen(path string) (string, error) {
	return c.DoAction("ws.open", path)
}

// WsList 工作区列表。
func (c *ClientFactory) WsList() (string, error) {
	return c.DoAction("ws.list", "")
}

// WsSwitch 切换工作区。
func (c *ClientFactory) WsSwitch(id string) (string, error) {
	return c.DoAction("ws.switch", id)
}

// WsClose 关闭工作区。
func (c *ClientFactory) WsClose(id string) (string, error) {
	return c.DoAction("ws.close", id)
}

// WsGen 工作区代数。
func (c *ClientFactory) WsGen() (int64, error) {
	text, err := c.DoAction("ws.gen", "")
	if err != nil {
		return 0, err
	}
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, "gen=") {
			n, e := strconv.ParseInt(strings.TrimSpace(line[4:]), 10, 64)
			return n, e
		}
	}
	return 0, nil
}

// DesignMeta 设计元数据。
func (c *ClientFactory) DesignMeta() (string, error) {
	return c.DoAction("design.meta", "")
}

// DesignGet 读设计属性。
func (c *ClientFactory) DesignGet(key string) (string, error) {
	return c.DoAction("design.get", key)
}

// DesignLease 写租约。
func (c *ClientFactory) DesignLease(token string) (string, error) {
	return c.DoAction("design.lease", token)
}

// DesignSet 租约写属性。
func (c *ClientFactory) DesignSet(key, value, token string) (string, error) {
	return c.DoAction("design.set", key+"="+value+"\n"+token)
}

// DesignRelease 释放租约。
func (c *ClientFactory) DesignRelease(token string) (string, error) {
	return c.DoAction("design.release", token)
}

// OpenBoard 打开工程（脚本模式返回路径占位，无 Arrow Board OO）。
func (c *ClientFactory) OpenBoard(path string) (string, error) {
	if err := c.ensure(); err != nil {
		return "", err
	}
	if strings.TrimSpace(path) == "" {
		return "", fmt.Errorf("empty board path")
	}
	if _, err := c.WsOpen(path); err != nil {
		return "", err
	}
	return path, nil
}

package hqeda

import (
	"fmt"
	"strings"
)

// ClientFactory 是 Flight IPC 的推荐产品入口（封装底层 FlightIPC）。
//
// 典型用法：
//
//	cli := hqeda.NewClientFactory()
//	if err := cli.Connect("", false); err != nil { ... }
//	defer cli.Disconnect()
//	board, err := cli.OpenBoard("/path/board.hqpcb")
//
// 低层逃生舱：NewFlightIPC / BindFlightIPC；应用代码优先用本类型。
type ClientFactory struct {
	*FlightIPC
}

// NewClientFactory 创建未连接的客户端工厂（产品默认构造入口）。
func NewClientFactory() *ClientFactory {
	return &ClientFactory{FlightIPC: NewFlightIPC()}
}

// ConnectService 连接后台服务（默认逻辑名 hqpcb-svc）。
func (c *ClientFactory) ConnectService(pipe string) error {
	if c == nil || c.FlightIPC == nil {
		return fmt.Errorf("nil ClientFactory")
	}
	return c.Connect(pipe, false)
}

// ConnectGUI 连接 GUI Interactive 端点（默认逻辑名 hqpcb-gui）。
func (c *ClientFactory) ConnectGUI(pipe string) error {
	if c == nil || c.FlightIPC == nil {
		return fmt.Errorf("nil ClientFactory")
	}
	return c.Connect(pipe, true)
}

// Transport 返回已绑定的 DoAction Transport（供 OO Session/Board 使用）。
func (c *ClientFactory) Transport() Transport {
	return BindFlightIPC(c.FlightIPC)
}

// DoGetFn 返回已绑定的 DoGet 回调（P6c）。
func (c *ClientFactory) DoGetFn() DoGetFn {
	return BindFlightDoGet(c.FlightIPC)
}

// Session 构造进程级 Session（握手 / config / Interactive）。
func (c *ClientFactory) Session() *Session {
	return NewSession(c.Transport())
}

// OpenBoard 打开工程并返回 Board（自动绑定 DoAction + DoGet）。
func (c *ClientFactory) OpenBoard(path string) (*Board, error) {
	if c == nil || c.FlightIPC == nil {
		return nil, fmt.Errorf("nil ClientFactory")
	}
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("empty board path")
	}
	return OpenBoard(c.Transport(), path, c.DoGetFn())
}

// WriteProps 持租约批量写设计属性（见 hqeda_write.go）。
func (c *ClientFactory) WriteProps(props map[string]string, token string) (map[string]string, error) {
	return WriteProps(c.FlightIPC, props, token)
}

// WriteConfig 写进程配置并读回。
func (c *ClientFactory) WriteConfig(key, value string) (string, error) {
	return WriteConfig(c.FlightIPC, key, value)
}

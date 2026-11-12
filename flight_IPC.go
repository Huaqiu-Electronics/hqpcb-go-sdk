// Package pcb 的本机 Arrow Flight IPC 客户端（P6d 权威路径）。
//
// 产品约定：只使用逻辑端点名（默认 hqpcb-svc），映射到本机 domain socket：
//
//	{tempdir}/{pipe}.flight.sock
//	→ grpc dial unix
//
// 不传 URI、不使用数字 TCP 端口。环境变量：
//
//	PCB_IPC_SERVICE_PIPE / PCB_IPC_GUI_PIPE — 覆盖默认逻辑名。
package pcb

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/flight"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// 默认逻辑端点名（与 C++ IPCPaths / Python 对齐）。
const (
	DefaultServicePipe = "hqpcb-svc"
	DefaultGUIPipe     = "hqpcb-gui"
)

// FlightCaps 为 hello/caps 行协议解析结果。
type FlightCaps struct {
	Interactive bool
	Workspace   bool
	Design      bool
	VerMajor    int
	VerMinor    int
	Raw         map[string]string
}

// FlightIPC 本机 Arrow Flight 客户端（按逻辑管道名连接）。
type FlightIPC struct {
	cli      flight.Client
	pipe     string
	peerRole string
	peerCaps FlightCaps
}

// SockPath 将逻辑名映射为本机 domain socket 绝对路径。
func SockPath(pipeName string) string {
	name := strings.TrimSpace(pipeName)
	if name == "" {
		name = DefaultServicePipe
	}
	return filepath.Join(os.TempDir(), name+".flight.sock")
}

// ResolvePipe 解析逻辑端点名：显式参数 > 环境变量 > 默认。
func ResolvePipe(pipeName string, gui bool) string {
	if s := strings.TrimSpace(pipeName); s != "" {
		return s
	}
	envKey := "PCB_IPC_SERVICE_PIPE"
	if gui {
		envKey = "PCB_IPC_GUI_PIPE"
	}
	if env := strings.TrimSpace(os.Getenv(envKey)); env != "" {
		return env
	}
	if gui {
		return DefaultGUIPipe
	}
	return DefaultServicePipe
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

// NewFlightIPC 创建未连接的客户端。
func NewFlightIPC() *FlightIPC {
	return &FlightIPC{peerCaps: FlightCaps{VerMajor: 1, Raw: map[string]string{}}}
}

// Pipe 返回当前逻辑端点名。
func (c *FlightIPC) Pipe() string { return c.pipe }

// PeerRole 返回 hello 角色。
func (c *FlightIPC) PeerRole() string { return c.peerRole }

// PeerCaps 返回 hello caps。
func (c *FlightIPC) PeerCaps() FlightCaps { return c.peerCaps }

// Connect 按逻辑端点名连接并 hello（默认 hqpcb-svc）。
func (c *FlightIPC) Connect(pipeName string, gui bool) error {
	c.Disconnect()
	name := ResolvePipe(pipeName, gui)
	sock := SockPath(name)
	if _, err := os.Stat(sock); err != nil {
		role := "service"
		if gui {
			role = "gui"
		}
		return fmt.Errorf("domain socket 不存在: %s\n请先启动：pcb_IPC_skel_serve --pipe %s --role %s", sock, name, role)
	}
	dialer := func(ctx context.Context, _ string) (net.Conn, error) {
		d := net.Dialer{Timeout: 5 * time.Second}
		return d.DialContext(ctx, "unix", sock)
	}
	// 大板 board.tracks / DoGet 可超过 gRPC 默认 4MiB。
	const maxMsg = 64 << 20
	cli, err := flight.NewClientWithMiddleware("passthrough:///hqpcb", nil, nil,
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(maxMsg),
			grpc.MaxCallSendMsgSize(maxMsg),
		),
	)
	if err != nil {
		return err
	}
	c.cli = cli
	c.pipe = name
	body, err := c.DoAction("hello", "")
	if err != nil {
		c.Disconnect()
		return err
	}
	caps := ParseCapsText(body)
	c.peerCaps = caps
	c.peerRole = caps.Raw["role"]
	if c.peerRole == "" {
		c.peerRole = "service"
	}
	return nil
}

// Disconnect 关闭客户端。
func (c *FlightIPC) Disconnect() {
	if c.cli != nil {
		_ = c.cli.Close()
		c.cli = nil
	}
	c.pipe = ""
	c.peerRole = "service"
	c.peerCaps = FlightCaps{VerMajor: 1, Raw: map[string]string{}}
}

// DoAction 调用命名 DoAction，返回首个 Result 正文（UTF-8）。
func (c *FlightIPC) DoAction(method, body string) (string, error) {
	if c.cli == nil {
		return "", fmt.Errorf("not connected")
	}
	if strings.TrimSpace(method) == "" {
		return "", fmt.Errorf("empty method")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	stream, err := c.cli.DoAction(ctx, &flight.Action{Type: method, Body: []byte(body)})
	if err != nil {
		return "", err
	}
	res, err := stream.Recv()
	if err == io.EOF {
		return "", fmt.Errorf("empty result for action=%s", method)
	}
	if err != nil {
		return "", err
	}
	// 排空剩余结果（本产品通常仅一条）。
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
	}
	return string(res.Body), nil
}

// Ping 探活回显。
func (c *FlightIPC) Ping(token string) (string, error) {
	if token == "" {
		token = "go"
	}
	return c.DoAction("ping", token)
}

// Caps 再次查询 caps。
func (c *FlightIPC) Caps() (FlightCaps, error) {
	text, err := c.DoAction("caps", "")
	if err != nil {
		return FlightCaps{}, err
	}
	caps := ParseCapsText(text)
	c.peerCaps = caps
	if role := caps.Raw["role"]; role != "" {
		c.peerRole = role
	}
	return caps, nil
}

// MethodsList 返回方法目录行协议。
func (c *FlightIPC) MethodsList() (string, error) {
	return c.DoAction("methods.list", "")
}

// ConfigGet 读进程配置。
func (c *FlightIPC) ConfigGet(key string) (string, error) {
	return c.DoAction("config.get", key)
}

// ConfigSet 写进程配置。
func (c *FlightIPC) ConfigSet(key, value string) (string, error) {
	return c.DoAction("config.set", key+"="+value)
}

// WsOpen 权威侧打开工程路径。
func (c *FlightIPC) WsOpen(path string) (string, error) {
	return c.DoAction("ws.open", path)
}

// WsGen 工作区代数。
func (c *FlightIPC) WsGen() (int64, error) {
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
	return 0, fmt.Errorf("bad ws.gen: %q", text)
}

// DesignLease 获取写租约。
func (c *FlightIPC) DesignLease(token string) (string, error) {
	return c.DoAction("design.lease", token)
}

// DesignSet 租约写设计属性。
func (c *FlightIPC) DesignSet(key, value, token string) (string, error) {
	return c.DoAction("design.set", key+"="+value+"\n"+token)
}

// DesignGet 读设计属性。
func (c *FlightIPC) DesignGet(key string) (string, error) {
	return c.DoAction("design.get", key)
}

// DesignRelease 释放写租约。
func (c *FlightIPC) DesignRelease(token string) (string, error) {
	return c.DoAction("design.release", token)
}

// DoGet 拉取 Ticket 对应 Arrow 表（P6c）。
func (c *FlightIPC) DoGet(ticket string) (arrow.Table, error) {
	if c.cli == nil {
		return nil, fmt.Errorf("not connected")
	}
	if strings.TrimSpace(ticket) == "" {
		return nil, fmt.Errorf("empty ticket")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	stream, err := c.cli.DoGet(ctx, &flight.Ticket{Ticket: []byte(ticket)})
	if err != nil {
		return nil, err
	}
	rdr, err := flight.NewRecordReader(stream)
	if err != nil {
		return nil, err
	}
	defer rdr.Release()
	recs := make([]arrow.Record, 0, 4)
	for rdr.Next() {
		rec := rdr.Record()
		rec.Retain()
		recs = append(recs, rec)
	}
	if err := rdr.Err(); err != nil {
		for _, r := range recs {
			r.Release()
		}
		return nil, err
	}
	if len(recs) == 0 {
		// 空流：用 schema 构造 0 行表。
		schema := rdr.Schema()
		if schema == nil {
			return nil, fmt.Errorf("empty DoGet stream")
		}
		slices := make([][]arrow.Array, schema.NumFields())
		for i, f := range schema.Fields() {
			b := array.NewBuilder(memory.DefaultAllocator, f.Type)
			arr := b.NewArray()
			b.Release()
			slices[i] = []arrow.Array{arr}
		}
		tbl := array.NewTableFromSlice(schema, slices)
		for _, col := range slices {
			for _, a := range col {
				a.Release()
			}
		}
		return tbl, nil
	}
	tbl := array.NewTableFromRecords(recs[0].Schema(), recs)
	for _, r := range recs {
		r.Release()
	}
	return tbl, nil
}

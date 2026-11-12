// Package pcb 的本机 Arrow IPC Streaming 客户端骨架（P1b：hello/ping/caps）。
//
// Deprecated: P6d 起请改用 FlightIPC（flight_IPC.go）与生成 API（pcb_api_gen.go）。
// 本文件仅保留兼容；新冒烟脚本 IPC_p6d_smoke.sh 不再调用 Streaming。
package pcb

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/ipc"
	"github.com/apache/arrow-go/v18/arrow/memory"
)

// IPCCaps 为握手 / caps 返回的能力摘要。
type IPCCaps struct {
	Interactive bool
	VerMajor    int32
	VerMinor    int32
}

// IPCEnvelope 为单行 IPC 信封。
type IPCEnvelope struct {
	Kind        string
	ID          int64
	Method      string
	Status      int32
	Role        string
	VerMajor    int32
	VerMinor    int32
	Interactive bool
	Detail      string
}

// IPCClient 本机 Arrow IPC 客户端。
type IPCClient struct {
	conn     net.Conn
	nextID   int64
	peerRole string
	peerCaps IPCCaps
	ready    bool
	alloc    memory.Allocator
	schema   *arrow.Schema
}

// NewIPCClient 创建客户端（未连接）。
func NewIPCClient() *IPCClient {
	return &IPCClient{
		nextID:   1,
		alloc:    memory.DefaultAllocator,
		schema:   ipcEnvelopeSchema(),
		peerCaps: IPCCaps{VerMajor: 1, VerMinor: 0},
	}
}

func ipcEnvelopeSchema() *arrow.Schema {
	return arrow.NewSchema([]arrow.Field{
		{Name: "kind", Type: arrow.BinaryTypes.String},
		{Name: "id", Type: arrow.PrimitiveTypes.Int64},
		{Name: "method", Type: arrow.BinaryTypes.String},
		{Name: "status", Type: arrow.PrimitiveTypes.Int32},
		{Name: "role", Type: arrow.BinaryTypes.String},
		{Name: "ver_major", Type: arrow.PrimitiveTypes.Int32},
		{Name: "ver_minor", Type: arrow.PrimitiveTypes.Int32},
		{Name: "interactive", Type: arrow.FixedWidthTypes.Boolean},
		{Name: "detail", Type: arrow.BinaryTypes.String},
	}, nil)
}

// PeerRole 返回握手角色。
func (c *IPCClient) PeerRole() string { return c.peerRole }

// PeerCaps 返回握手 caps。
func (c *IPCClient) PeerCaps() IPCCaps { return c.peerCaps }

// Connect 连接 Unix 套接字路径并完成 hello。
func (c *IPCClient) Connect(path string, timeout time.Duration) error {
	c.Disconnect()
	dialer := net.Dialer{Timeout: timeout}
	conn, err := dialer.Dial("unix", path)
	if err != nil {
		return err
	}
	c.conn = conn
	_ = conn.SetDeadline(time.Now().Add(timeout))
	hello := IPCEnvelope{Kind: "hello", Role: "client", VerMajor: 1, VerMinor: 0}
	if err := c.writeEnv(hello); err != nil {
		c.Disconnect()
		return err
	}
	ack, err := c.readEnv()
	if err != nil {
		c.Disconnect()
		return err
	}
	if ack.Kind != "hello_ack" || ack.Status != 0 || ack.VerMajor != 1 {
		c.Disconnect()
		return fmt.Errorf("hello rejected: %+v", ack)
	}
	c.peerRole = ack.Role
	c.peerCaps = IPCCaps{Interactive: ack.Interactive, VerMajor: ack.VerMajor, VerMinor: ack.VerMinor}
	c.ready = true
	_ = conn.SetDeadline(time.Time{})
	return nil
}

// Disconnect 关闭连接。
func (c *IPCClient) Disconnect() {
	c.ready = false
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
	}
}

// Ping 调用 ping。
func (c *IPCClient) Ping(token string, timeout time.Duration) (string, error) {
	resp, err := c.roundTrip("ping", token, timeout)
	if err != nil {
		return "", err
	}
	if resp.Status != 0 {
		return "", fmt.Errorf("ping status=%d", resp.Status)
	}
	return resp.Detail, nil
}

// Caps 调用 caps。
func (c *IPCClient) Caps(timeout time.Duration) (IPCCaps, error) {
	resp, err := c.roundTrip("caps", "", timeout)
	if err != nil {
		return IPCCaps{}, err
	}
	if resp.Status != 0 {
		return IPCCaps{}, fmt.Errorf("caps status=%d", resp.Status)
	}
	return IPCCaps{Interactive: resp.Interactive, VerMajor: resp.VerMajor, VerMinor: resp.VerMinor}, nil
}

func (c *IPCClient) roundTrip(method, detail string, timeout time.Duration) (IPCEnvelope, error) {
	if !c.ready || c.conn == nil {
		return IPCEnvelope{}, fmt.Errorf("not connected")
	}
	_ = c.conn.SetDeadline(time.Now().Add(timeout))
	defer func() { _ = c.conn.SetDeadline(time.Time{}) }()
	id := c.nextID
	c.nextID++
	req := IPCEnvelope{Kind: "req", ID: id, Method: method, Detail: detail}
	if err := c.writeEnv(req); err != nil {
		return IPCEnvelope{}, err
	}
	resp, err := c.readEnv()
	if err != nil {
		return IPCEnvelope{}, err
	}
	if resp.Kind != "resp" || resp.ID != id || resp.Method != method {
		return IPCEnvelope{}, fmt.Errorf("bad response: %+v", resp)
	}
	return resp, nil
}

func (c *IPCClient) writeEnv(env IPCEnvelope) error {
	payload, err := encodeIPCEnvelope(c.alloc, c.schema, env)
	if err != nil {
		return err
	}
	return writeIPCFrame(c.conn, payload)
}

func (c *IPCClient) readEnv() (IPCEnvelope, error) {
	payload, err := readIPCFrame(c.conn)
	if err != nil {
		return IPCEnvelope{}, err
	}
	return decodeIPCEnvelope(c.alloc, c.schema, payload)
}

func encodeIPCEnvelope(alloc memory.Allocator, schema *arrow.Schema, env IPCEnvelope) ([]byte, error) {
	b := array.NewRecordBuilder(alloc, schema)
	defer b.Release()
	b.Field(0).(*array.StringBuilder).Append(env.Kind)
	b.Field(1).(*array.Int64Builder).Append(env.ID)
	b.Field(2).(*array.StringBuilder).Append(env.Method)
	b.Field(3).(*array.Int32Builder).Append(env.Status)
	b.Field(4).(*array.StringBuilder).Append(env.Role)
	b.Field(5).(*array.Int32Builder).Append(env.VerMajor)
	b.Field(6).(*array.Int32Builder).Append(env.VerMinor)
	b.Field(7).(*array.BooleanBuilder).Append(env.Interactive)
	b.Field(8).(*array.StringBuilder).Append(env.Detail)
	rec := b.NewRecordBatch()
	defer rec.Release()
	payload, err := ipc.GetRecordBatchPayload(rec, ipc.WithAllocator(alloc))
	if err != nil {
		return nil, err
	}
	defer payload.Release()
	var buf bytes.Buffer
	if _, err := payload.WritePayload(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func decodeIPCEnvelope(alloc memory.Allocator, schema *arrow.Schema, payload []byte) (IPCEnvelope, error) {
	// 宿主载荷是单条 RecordBatch Message；前置 Schema Message 组成可读 stream。
	sp := ipc.GetSchemaPayload(schema, alloc)
	defer sp.Release()
	var stream bytes.Buffer
	if _, err := sp.WritePayload(&stream); err != nil {
		return IPCEnvelope{}, err
	}
	if _, err := stream.Write(payload); err != nil {
		return IPCEnvelope{}, err
	}
	r, err := ipc.NewReader(bytes.NewReader(stream.Bytes()), ipc.WithAllocator(alloc))
	if err != nil {
		return IPCEnvelope{}, err
	}
	defer r.Release()
	rec, err := r.Read()
	if err != nil {
		return IPCEnvelope{}, err
	}
	defer rec.Release()
	if rec.NumRows() < 1 {
		return IPCEnvelope{}, fmt.Errorf("empty envelope")
	}
	return IPCEnvelope{
		Kind:        rec.Column(0).(*array.String).Value(0),
		ID:          rec.Column(1).(*array.Int64).Value(0),
		Method:      rec.Column(2).(*array.String).Value(0),
		Status:      rec.Column(3).(*array.Int32).Value(0),
		Role:        rec.Column(4).(*array.String).Value(0),
		VerMajor:    rec.Column(5).(*array.Int32).Value(0),
		VerMinor:    rec.Column(6).(*array.Int32).Value(0),
		Interactive: rec.Column(7).(*array.Boolean).Value(0),
		Detail:      rec.Column(8).(*array.String).Value(0),
	}, nil
}

func writeIPCFrame(w io.Writer, payload []byte) error {
	var hdr [4]byte
	binary.BigEndian.PutUint32(hdr[:], uint32(len(payload)))
	if _, err := w.Write(hdr[:]); err != nil {
		return err
	}
	_, err := w.Write(payload)
	return err
}

func readIPCFrame(r io.Reader) ([]byte, error) {
	var hdr [4]byte
	if _, err := io.ReadFull(r, hdr[:]); err != nil {
		return nil, err
	}
	n := binary.BigEndian.Uint32(hdr[:])
	if n > 64*1024*1024 {
		return nil, fmt.Errorf("frame too large")
	}
	buf := make([]byte, n)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

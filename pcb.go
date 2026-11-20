// Package hqeda 是 PCB_V5 TinyGo/WASI 脚本的宿主桥接 SDK。
//
// 通过共享目录上的 JSON 请求/响应与桌面宿主通信；不直接访问设计内核。
// 环境变量 PCB_HOST_DIR 指向宿主创建的临时目录（含 req/ 与 resp/）。
//
// 原始桥：Log / Cmdlet / Query。
// 类型化 cmdlet：Workspace* / Project* / Script* / Plugins / Window* / Help / LoadPlugin
//（见 workspace.go、project.go、script.go、app.go）；
// Query 结构化：QueryTableResult / QueryMaps / QueryTable.Cell|RowMap|Maps（见 query.go）；
// 常用 runtime 表：QueryPlugins / QueryLayers / QueryViewOptions / QueryCmdletHist /
// QueryCmdHist / QueryConfigItems（见 query_rt.go）；
// 命令行拼装：EncodeValue / KvToken / Line（见 line.go），对齐宿主 CmdletKvArgs。
package hqeda

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// hostReq 是写入 req/ 的请求体。
type hostReq struct {
	Op   string                 `json:"op"`
	Args map[string]interface{} `json:"args"`
}

// hostResp 是宿主写回 resp/ 的响应体。
type hostResp struct {
	OK     bool   `json:"ok"`
	Result string `json:"result"`
	Error  string `json:"error"`
}

// hostDir 返回宿主约定的共享根目录。
func hostDir() (string, error) {
	dir := os.Getenv("PCB_HOST_DIR")
	if dir == "" {
		return "", errors.New("pcb: PCB_HOST_DIR not set")
	}
	return dir, nil
}

// call 发送一次 host 请求并等待响应（轮询，最长约 120s）。
func call(op string, args map[string]interface{}) (string, error) {
	return callFor(op, args, 120*time.Second)
}

// callFor 同 call，但可指定最长等待时间（向导 wait.* 用更长超时）。
func callFor(op string, args map[string]interface{}, maxWait time.Duration) (string, error) {
	root, err := hostDir()
	if err != nil {
		return "", err
	}
	id := fmt.Sprintf("%d", time.Now().UnixNano())
	reqPath := filepath.Join(root, "req", id+".json")
	respPath := filepath.Join(root, "resp", id+".json")
	body, err := json.Marshal(hostReq{Op: op, Args: args})
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(reqPath, body, 0o600); err != nil {
		return "", err
	}
	if maxWait <= 0 {
		maxWait = 120 * time.Second
	}
	deadline := time.Now().Add(maxWait)
	for time.Now().Before(deadline) {
		data, err := os.ReadFile(respPath)
		if err != nil {
			time.Sleep(10 * time.Millisecond)
			continue
		}
		// 空/半残响应：等宿主 *.part rename 完成后再解析。
		if len(bytes.TrimSpace(data)) == 0 {
			time.Sleep(10 * time.Millisecond)
			continue
		}
		var resp hostResp
		if err := json.Unmarshal(data, &resp); err != nil {
			time.Sleep(10 * time.Millisecond)
			continue
		}
		_ = os.Remove(reqPath)
		_ = os.Remove(respPath)
		if !resp.OK {
			if resp.Error == "" {
				return "", errors.New("pcb: host call failed")
			}
			return "", errors.New(resp.Error)
		}
		return resp.Result, nil
	}
	return "", errors.New("pcb: host call timeout")
}

// Log 将消息发往宿主输出（不经过 cmdlet）。
func Log(msg string) {
	_, _ = call("log", map[string]interface{}{"message": msg})
}

// Cmdlet 执行一条宿主 cmdlet 命令行，返回结果文本。
func Cmdlet(line string) (string, error) {
	return call("cmdlet", map[string]interface{}{"line": line})
}

// ResolveFile 经宿主 path.resolve 把相对路径解析为已存在绝对路径。
func ResolveFile(rel string) (string, error) {
	rel = strings.TrimSpace(strings.ReplaceAll(rel, "\\", "/"))
	if rel == "" {
		return "", fmt.Errorf("empty path")
	}
	out, err := CmdletLine("path.resolve", KvToken("path", rel))
	if err != nil {
		return "", err
	}
	out = strings.TrimSpace(out)
	if out == "" {
		return "", fmt.Errorf("path not found: %s", rel)
	}
	return out, nil
}

// SleepMs 经宿主 sleep.ms 停顿（会 processEvents，供 Route 等交互脚本跟手刷新）。
// ms < 0 时按 0 处理。
func SleepMs(ms int) error {
	if ms < 0 {
		ms = 0
	}
	_, err := Cmdlet(fmt.Sprintf("sleep.ms %d", ms))
	return err
}

// SelectLayer 经宿主 layer_schema.select 切换当前层（层名或层 id）。
func SelectLayer(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("empty layer name")
	}
	_, err := CmdletLine("layer_schema.select", name)
	return err
}

// WaitOpenFollow 等待场景跟随打开完成（rasternext.wait_open_follow）。
// timeoutMs < 0 时按 0 处理。
func WaitOpenFollow(timeoutMs int) error {
	if timeoutMs < 0 {
		timeoutMs = 0
	}
	_, err := Cmdlet(fmt.Sprintf("rasternext.wait_open_follow %d", timeoutMs))
	return err
}

// ZoomAll 执行 zoomall（全图适配）。
func ZoomAll() error {
	_, err := Cmdlet("zoomall")
	return err
}

// StartAction 启动命名工具动作（action.start <name>）。
func StartAction(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("empty action name")
	}
	_, err := CmdletLine("action.start", name)
	return err
}

// CurrentAction 查询当前动作名（action.current）。
func CurrentAction() (string, error) {
	return Cmdlet("action.current")
}

// MouseMoveEx 直线插值跟手（action.input move-line）。
func MouseMoveEx(x0, y0, x1, y1 float64, steps int) error {
	_, err := Cmdlet(fmt.Sprintf("action.input move-line %g %g %g %g %d", x0, y0, x1, y1, steps))
	return err
}

// MouseMove 移动指针到纸面坐标（action.input move）。
func MouseMove(x, y float64) error {
	_, err := Cmdlet(fmt.Sprintf("action.input move %g %g", x, y))
	return err
}

// MouseClick 单击纸面坐标（action.input click）。
func MouseClick(x, y float64) error {
	_, err := Cmdlet(fmt.Sprintf("action.input click %g %g", x, y))
	return err
}

// MouseDBClick 双击纸面坐标（action.input dbclick）。
func MouseDBClick(x, y float64) error {
	_, err := Cmdlet(fmt.Sprintf("action.input dbclick %g %g", x, y))
	return err
}

// CancelAction 取消当前工具动作（action.cancel）。
func CancelAction() error {
	_, err := Cmdlet("action.cancel")
	return err
}

// Query 执行一条宿主 SQL（与 runtime.query / telnet 同源策略），返回 JSON 文本。
func Query(sql string) (string, error) {
	return call("query", map[string]interface{}{"sql": sql})
}

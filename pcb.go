// Package hqeda 是 PCB_V5 TinyGo/WASI 脚本的宿主桥接 SDK。
//
// 缺省（无 -tags witbridge）：共享目录 JSON 请求/响应（PCB_HOST_DIR req/resp）。
// witbridge：经 hqpcb:host/op WIT import（需 wasip2 + adapter 组合或嵌入 linker）。
// 不直接访问设计内核。
//
// 原始桥：Log / Cmdlet / Query。
// 类型化 cmdlet：Workspace* / Project* / Script* / Plugins / Window* / Help / LoadPlugin
// （见 workspace.go、project.go、script.go、app.go）；
// Query 结构化：QueryTableResult / QueryMaps / QueryTable.Cell|RowMap|Maps（见 query.go）；
// 常用 runtime 表：QueryPlugins / QueryLayers / QueryViewOptions / QueryCmdletHist /
// QueryCmdHist / QueryConfigItems（见 query_rt.go）；
// 命令行拼装：EncodeValue / KvToken / Line（见 line.go），对齐宿主 CmdletKvArgs。
package hqeda

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// hostDir 返回宿主约定的共享根目录。
func hostDir() (string, error) {
	dir := os.Getenv("PCB_HOST_DIR")
	if dir == "" {
		return "", errors.New("pcb: PCB_HOST_DIR not set")
	}
	return dir, nil
}

// hostCallDepth 统计嵌套中的 Guest→Host call 层数。
//
// call 等待 resp 时会泵 inbox；深度>0 时跳过 MouseMove，避免 ApplyMeasurePrev 重入叠栈卡死。
var hostCallDepth int

// Log 将消息发往宿主输出（不经过 cmdlet）。
// 有额外参数时按 fmt.Sprintf(format, args...) 格式化；无参则原样输出 format。
func Log(format string, args ...interface{}) {
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
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

// SetActionParam 设置当前动作参数（action.setParam；默认发往 Command MP）。
// name 为 HQML focus-target（如 angleLock / lineWidth）；value 为下标或选项文案。
func SetActionParam(name string, value interface{}) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("empty action param name")
	}
	// nil 先拦，避免 fmt.Sprint(nil) 变成 "<nil>"。
	if value == nil {
		return "", fmt.Errorf("empty action param value")
	}
	v := strings.TrimSpace(fmt.Sprint(value))
	if v == "" {
		return "", fmt.Errorf("empty action param value")
	}
	return CmdletLine(
		"action.setParam",
		KvToken("name", name),
		KvToken("value", v),
	)
}

// ChooseFile 弹出工程打开对话框，返回所选绝对路径（不自动开板）。
func ChooseFile() (string, error) {
	out, err := Cmdlet("file.choose")
	if err != nil {
		return "", err
	}
	out = strings.TrimSpace(out)
	if out == "" {
		return "", fmt.Errorf("chooseFile cancelled")
	}
	return out, nil
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

// MouseBoxSelect 对角框选拖拽（action.input box-select）。
// steps<=0 时按 8 步插值。
func MouseBoxSelect(x0, y0, x1, y1 float64, steps int) error {
	if steps <= 0 {
		steps = 8
	}
	_, err := Cmdlet(fmt.Sprintf("action.input box-select %g %g %g %g %d", x0, y0, x1, y1, steps))
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

// MousePress 左键按下（action.input press）。
func MousePress(x, y float64) error {
	_, err := Cmdlet(fmt.Sprintf("action.input press %g %g", x, y))
	return err
}

// MouseRelease 左键抬起（action.input release）。
func MouseRelease(x, y float64) error {
	_, err := Cmdlet(fmt.Sprintf("action.input release %g %g", x, y))
	return err
}

// CancelAction 取消当前工具动作（action.cancel）。
func CancelAction() error {
	_, err := Cmdlet("action.cancel")
	return err
}

// PickAt 文档 mm 拾取并选中（rasternext.pick），返回命中业务 id。
func PickAt(x, y float64) ([]int64, error) {
	out, err := Cmdlet(fmt.Sprintf("rasternext.pick %g %g", x, y))
	if err != nil {
		return nil, err
	}
	const prefix = "ids="
	idx := strings.Index(out, prefix)
	if idx < 0 {
		return nil, nil
	}
	raw := strings.TrimSpace(out[idx+len(prefix):])
	if i := strings.IndexAny(raw, " \t\r\n"); i >= 0 {
		raw = raw[:i]
	}
	if raw == "" {
		return nil, nil
	}
	parts := strings.Split(raw, ",")
	ids := make([]int64, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n, e := strconv.ParseInt(p, 10, 64)
		if e != nil || n <= 0 {
			continue
		}
		ids = append(ids, n)
	}
	return ids, nil
}

// HideItems 隐藏对象（rasternext.hide_items → setStaticObjectsVisibleByIds）。
// ids 为空时隐藏当前 selection。
func HideItems(ids []int64) (string, error) {
	if len(ids) == 0 {
		return Cmdlet("rasternext.hide_items selection")
	}
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		parts = append(parts, strconv.FormatInt(id, 10))
	}
	return Cmdlet("rasternext.hide_items " + strings.Join(parts, ","))
}

// Query 执行一条宿主 SQL（与 runtime.query / telnet 同源策略），返回 JSON 文本。
func Query(sql string) (string, error) {
	return call("query", map[string]interface{}{"sql": sql})
}

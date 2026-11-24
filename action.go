// Go WASM 插件交互动作：登记 HostedEdaAction + 测量叠加（apply_measure_prev）。
//
// 鼠标 / 按键经宿主 inbox op=action.event 转发；Guest 通过 ApplyMeasurePrev
// 更新 RnViewAreaPrev，不直接访问画布内核。
package hqeda

import (
	"fmt"
	"strings"
)

// PointMm 表示文档坐标系下的毫米点（与宿主 paper_x / paper_y 同单位）。
type PointMm struct {
	// X 水平坐标，向右为正，单位 mm。
	X float64 `json:"x"`
	// Y 垂直坐标，与宿主纸面约定一致，单位 mm。
	Y float64 `json:"y"`
}

// MeasureLabel 是测量叠加上的一处文字锚点（文档 mm）。
type MeasureLabel struct {
	// X 锚点横坐标，单位 mm。
	X float64 `json:"x"`
	// Y 锚点纵坐标，单位 mm。
	Y float64 `json:"y"`
	// Text 显示文案；空串时宿主跳过该标注。
	Text string `json:"text"`
}

// ActionEvent 是宿主转发的一次输入事件。
type ActionEvent struct {
	// Name 动作命令 id（与 RegisterAction 的 name 一致）。
	Name string
	// Type 事件类型名（如 MouseMove、MousePress、KeyPress）。
	Type string
	// Payload 事件负载（含 paper_x、paper_y、button、key 等）。
	Payload map[string]interface{}
}

// ActionHandler 是插件侧交互动作回调；由 HostedEdaAction 经 inbox 转发。
type ActionHandler interface {
	// OnEvent 处理一次输入事件。
	//
	// consumed 为 true 表示已消费；err 非 nil 时宿主可记失败。
	OnEvent(ev ActionEvent) (consumed bool, err error)
	// OnStop 在动作停止 / 打断时清理预览与内部状态。
	//
	// name 为动作命令 id。
	OnStop(name string) error
}

// actionHandlers 以动作 name 索引本地 ActionHandler。
var actionHandlers = map[string]ActionHandler{}

// HandleAction 登记本地动作处理。
//
// name 须与随后 RegisterAction 的 name 一致；h 为 nil 或 name 为空则忽略。
func HandleAction(name string, h ActionHandler) {
	name = strings.TrimSpace(name)
	if name == "" || h == nil {
		return
	}
	actionHandlers[name] = h
}

// RegisterAction 向宿主工作区登记动作命令。
//
// name 为 cmdlet / 动作 id；shortName / fullName / displayName 供 UI 与 HUD。
// 须在 onWorkspaceInit（或工作区已激活）后调用；失败返回桥错误。
func RegisterAction(name, shortName, fullName, displayName string) error {
	_, err := call("register_action", map[string]interface{}{
		"name":        name,
		"shortName":   shortName,
		"fullName":    fullName,
		"displayName": displayName,
	})
	return err
}

// ApplyMeasurePrev 更新尺子类测量橡皮筋（轮廓 + 标注）。
//
// outline 为折线顶点（建议 ≥2 点）；labels 可空；fontSizeMm≤0 时宿主用默认字号。
// 成功返回 nil；宿主无画布或桥失败时返回 error。
func ApplyMeasurePrev(outline []PointMm, labels []MeasureLabel, fontSizeMm float64) error {
	// 转成 JSON 友好的 map 切片，避免结构体标签在 TinyGo 侧差异。
	out := make([]map[string]interface{}, 0, len(outline))
	for _, p := range outline {
		out = append(out, map[string]interface{}{"x": p.X, "y": p.Y})
	}
	labs := make([]map[string]interface{}, 0, len(labels))
	for _, l := range labels {
		labs = append(labs, map[string]interface{}{"x": l.X, "y": l.Y, "text": l.Text})
	}
	_, err := call("apply_measure_prev", map[string]interface{}{
		"outline":    out,
		"labels":     labs,
		"fontSizeMm": fontSizeMm,
	})
	return err
}

// ClearMeasurePrev 清除测量叠加层。
//
// 成功返回 nil；无 sink 时宿主仍可返回成功。
func ClearMeasurePrev() error {
	_, err := call("clear_measure_prev", map[string]interface{}{})
	return err
}

// dispatchActionEvent 将 inbox action.event 分派到已登记 ActionHandler。
//
// args 须含 name、type，可选 payload。返回 "1"/"0" 表示是否消费；无 handler 时 "0"。
func dispatchActionEvent(args map[string]interface{}) (string, error) {
	name, _ := args["name"].(string)
	typ, _ := args["type"].(string)
	name = strings.TrimSpace(name)
	h := actionHandlers[name]
	if h == nil {
		return "0", nil
	}
	payload, _ := args["payload"].(map[string]interface{})
	if payload == nil {
		payload = map[string]interface{}{}
	}
	consumed, err := h.OnEvent(ActionEvent{Name: name, Type: typ, Payload: payload})
	if err != nil {
		return "", err
	}
	if consumed {
		return "1", nil
	}
	return "0", nil
}

// dispatchActionStop 将 inbox action.stop 分派到 ActionHandler.OnStop。
//
// args 须含 name；无 handler 时视为成功返回 "ok"。
func dispatchActionStop(args map[string]interface{}) (string, error) {
	name, _ := args["name"].(string)
	name = strings.TrimSpace(name)
	h := actionHandlers[name]
	if h == nil {
		return "ok", nil
	}
	if err := h.OnStop(name); err != nil {
		return "", err
	}
	return "ok", nil
}

// PayloadFloat 从事件 payload 取浮点数字段。
//
// payload 可 nil；key 缺失或无法解析时返回 0。
func PayloadFloat(payload map[string]interface{}, key string) float64 {
	if payload == nil {
		return 0
	}
	v, ok := payload[key]
	if !ok || v == nil {
		return 0
	}
	switch t := v.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case int:
		return float64(t)
	case int64:
		return float64(t)
	default:
		// 字符串等其它 JSON 形态尽力扫描。
		var f float64
		_, _ = fmt.Sscan(fmt.Sprint(t), &f)
		return f
	}
}

// PayloadInt 从事件 payload 取整型字段（经 PayloadFloat 截断）。
//
// payload 可 nil；key 缺失时返回 0。
func PayloadInt(payload map[string]interface{}, key string) int {
	return int(PayloadFloat(payload, key))
}

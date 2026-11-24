// 常驻 Go WASM 插件：生命周期事件循环 + 命令 / 面板 / 数据总线 API。
//
// 仅当宿主设置 PCB_PLUGIN_MODE=1（GoWasmResident）时 Run 才有意义；
// 普通 script.go.run 一次性脚本勿调用 Run。Host→Guest 走 inbox/inbox_ack，
// Guest→Host 仍走 req/resp（与 pcb.go 的 call 相同）。
package hqeda

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LifecycleHandler 处理 Host→Guest 生命周期事件（无参；失败返回 error）。
type LifecycleHandler func() error

// CommandHandler 处理宿主命令回调（如菜单 builtin.ruler.open）。
//
// name 为登记时的命令 id，与 RegisterCommand 的 name 一致。
type CommandHandler func(name string) error

// HqcHandler 处理小程序 hqc 方法（args 为 JSON 对象，可能为空 map）。
type HqcHandler func(args map[string]interface{}) error

var (
	// lifecycleHandlers 以短名（如 onWndInit）索引生命周期回调。
	lifecycleHandlers = map[string]LifecycleHandler{}
	// commandHandlers 以命令 id 索引本地命令处理。
	commandHandlers = map[string]CommandHandler{}
	// hqcHandlers 以 hqc 方法名（如 applyOptions）索引面板回调。
	hqcHandlers = map[string]HqcHandler{}
	// pluginStop 为 true 时 Run 退出事件循环。
	pluginStop = false
)

// OnLifecycle 登记生命周期回调。
//
// name 不含 "lifecycle." 前缀（如 "onWndInit"）；h 为 nil 或 name 为空则忽略。
func OnLifecycle(name string, h LifecycleHandler) {
	// 规范化名称，避免空白键污染表。
	name = strings.TrimSpace(name)
	if name == "" || h == nil {
		return
	}
	lifecycleHandlers[name] = h
}

// HandleCommand 登记本地命令处理（在收到 command.invoke 时调用）。
//
// name 须与随后 RegisterCommand 使用的 id 一致；h 为 nil 则忽略。
func HandleCommand(name string, h CommandHandler) {
	name = strings.TrimSpace(name)
	if name == "" || h == nil {
		return
	}
	commandHandlers[name] = h
}

// HandleHqc 登记 hqc 方法（如 applyOptions / clearResults）。
//
// name 为页面 hqc.<name> 中的方法名；宿主经 inbox op "hqc.<name>" 转发。
func HandleHqc(name string, h HqcHandler) {
	name = strings.TrimSpace(name)
	if name == "" || h == nil {
		return
	}
	hqcHandlers[name] = h
}

// RegisterCommand 向宿主登记一条可被菜单 / cmdlet 触发的命令。
//
// name 为稳定命令 id；shortName / fullName 供 UI 展示，可空（宿主回退为 name）。
// 须在生命周期回调（如 onWndInit）内调用；失败时返回宿主桥错误。
func RegisterCommand(name, shortName, fullName string) error {
	// 经文件桥登记；宿主侧绑 Promise 回调再投递 command.invoke。
	_, err := call("register_command", map[string]interface{}{
		"name":      name,
		"shortName": shortName,
		"fullName":  fullName,
	})
	return err
}

// ShowMpPanel 打开 HQML 命令面板页面。
//
// bundleId 对应 plugin.yaml 的 pages 键（如 "ruler"）。
// 成功返回对话框 id 字符串；失败时 error 非 nil。
func ShowMpPanel(bundleId string) (string, error) {
	return call("show_mp_panel", map[string]interface{}{"bundleId": bundleId})
}

// PostDataBus 向宿主数据总线投递 topic。
//
// topic 为总线主题；payload 可 nil（视为空对象）。失败返回桥错误。
func PostDataBus(topic string, payload map[string]interface{}) error {
	// nil payload 统一成空对象，避免宿主 JSON 缺字段。
	if payload == nil {
		payload = map[string]interface{}{}
	}
	_, err := call("post_data_bus", map[string]interface{}{
		"topic":   topic,
		"payload": payload,
	})
	return err
}

// ShowMessage 显示状态栏 / 提示消息。
//
// text 为展示文案；level 为严重级别（与宿主约定，常用 0）；timeoutMs 为超时毫秒。
func ShowMessage(text string, level, timeoutMs int) error {
	_, err := call("show_message", map[string]interface{}{
		"text":      text,
		"level":     level,
		"timeoutMs": timeoutMs,
	})
	return err
}

// ExecuteCmdletLine 经插件桥执行一条完整 cmdlet 命令行。
//
// line 与 Cmdlet 同语义；成功返回 nil，失败返回宿主诊断。
func ExecuteCmdletLine(line string) error {
	_, err := call("execute_cmdlet", map[string]interface{}{"line": line})
	return err
}

// PluginMode 报告当前是否处于常驻插件模式（PCB_PLUGIN_MODE=1）。
func PluginMode() bool {
	return os.Getenv("PCB_PLUGIN_MODE") == "1"
}

// Run 进入常驻事件循环，直到收到 lifecycle.shutdown。
//
// 非插件模式下仅打日志并返回；须已设置 PCB_HOST_DIR。
func Run() {
	// 一次性脚本误调 Run 时直接退出，避免空转。
	if !PluginMode() {
		Log("pcb: Run() called outside plugin mode; exiting")
		return
	}
	root, err := hostDir()
	if err != nil {
		Log("pcb: %v", err)
		return
	}
	// 确保 Host→Guest 目录存在（宿主通常已建，Guest 再兜底）。
	inbox := filepath.Join(root, "inbox")
	_ = os.MkdirAll(inbox, 0o700)
	_ = os.MkdirAll(filepath.Join(root, "inbox_ack"), 0o700)

	for !pluginStop {
		entries, err := os.ReadDir(inbox)
		if err != nil {
			// 目录瞬态不可读时退避，避免忙等。
			time.Sleep(10 * time.Millisecond)
			continue
		}
		handled := false
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
				continue
			}
			path := filepath.Join(inbox, e.Name())
			// 文件名（无扩展名）即事件 id，与宿主 postInbox 约定一致。
			id := strings.TrimSuffix(e.Name(), ".json")
			if processInboxFile(path, id) {
				handled = true
			}
		}
		if !handled {
			time.Sleep(10 * time.Millisecond)
		}
	}
}

// inboxEnvelope 是 Host→Guest 单条 inbox 事件的 JSON 形状。
type inboxEnvelope struct {
	// Op 事件名（如 lifecycle.onWndInit、action.event）。
	Op string `json:"op"`
	// Args 事件参数对象；可空。
	Args map[string]interface{} `json:"args"`
}

// processInboxFile 读取并分派单个 inbox 文件，写 ack 后删除请求。
//
// path 为 inbox 下 json 绝对路径；id 为无扩展名的事件 id。
// 成功处理返回 true；半残/解析失败返回 false（保留文件供重试）。
func processInboxFile(path, id string) bool {
	data, err := os.ReadFile(path)
	// 空文件视为宿主尚未写完，本轮跳过。
	if err != nil || len(strings.TrimSpace(string(data))) == 0 {
		return false
	}
	var env inboxEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return false
	}
	result, callErr := dispatchInbox(env.Op, env.Args)
	ack := map[string]interface{}{"ok": callErr == nil, "result": result}
	if callErr != nil {
		ack["error"] = callErr.Error()
		ack["ok"] = false
	}
	writeInboxAck(id, ack)
	_ = os.Remove(path)
	return true
}

// writeInboxAck 将 ack 原子写入 inbox_ack/<id>.json。
//
// id 与请求文件基名一致；ack 须含 ok，失败时含 error。
func writeInboxAck(id string, ack map[string]interface{}) {
	root, err := hostDir()
	if err != nil {
		return
	}
	body, err := json.Marshal(ack)
	if err != nil {
		return
	}
	ackPath := filepath.Join(root, "inbox_ack", id+".json")
	// 先写 .part 再 rename，避免宿主读到半残 JSON。
	tmp := ackPath + ".part"
	if err := os.WriteFile(tmp, body, 0o600); err != nil {
		return
	}
	_ = os.Rename(tmp, ackPath)
}

// dispatchInbox 按 op 分派到生命周期 / 命令 / 动作 / hqc 处理器。
//
// op 为完整事件名；args 为参数对象（可 nil）。
// 返回 result 文本与错误；未知 op 返回 error。
func dispatchInbox(op string, args map[string]interface{}) (string, error) {
	switch {
	case op == "lifecycle.shutdown":
		// 置位后 Run 循环下一轮退出。
		pluginStop = true
		return "ok", nil
	case strings.HasPrefix(op, "lifecycle."):
		name := strings.TrimPrefix(op, "lifecycle.")
		if h, ok := lifecycleHandlers[name]; ok && h != nil {
			if err := h(); err != nil {
				return "", err
			}
		}
		return "ok", nil
	case op == "command.invoke":
		name, _ := args["name"].(string)
		name = strings.TrimSpace(name)
		h := commandHandlers[name]
		if h == nil {
			return "", fmt.Errorf("no command handler: %s", name)
		}
		if err := h(name); err != nil {
			return "", err
		}
		return "ok", nil
	case op == "action.event":
		return dispatchActionEvent(args)
	case op == "action.stop":
		return dispatchActionStop(args)
	case strings.HasPrefix(op, "hqc."):
		name := strings.TrimPrefix(op, "hqc.")
		h := hqcHandlers[name]
		// 未登记的 hqc 视为成功，避免面板调用打挂插件。
		if h == nil {
			return "ok", nil
		}
		if err := h(args); err != nil {
			return "", err
		}
		return "ok", nil
	default:
		return "", fmt.Errorf("unknown inbox op: %s", op)
	}
}

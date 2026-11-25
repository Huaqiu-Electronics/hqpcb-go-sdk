//go:build witbridge

package hqeda

import (
	"encoding/json"

	hostpull "github.com/Huaqiu-Electronics/hqpcb-go-sdk/internal/witgen/hqpcb/host/host-pull"
)

// inboxFileBridgeEnabled：witbridge 桌面轨只用 host-pull，不再扫 inbox 文件。
func inboxFileBridgeEnabled() bool { return false }

// pumpWitPullOnce 经 host-pull.take-event 取异步 Host→Guest 事件。
//
// 有事件时分派并 ack-event，返回 true；队列空返回 false（调用方休眠再 poll）。
func pumpWitPullOnce() bool {
	handled := false
	for {
		opt := hostpull.TakeEvent()
		if opt.None() {
			break
		}
		tup := opt.Value()
		id, name, argsJSON := tup[0], tup[1], tup[2]
		if hostCallDepth > 0 && isActionMouseMoveJSON(argsJSON) {
			// 嵌套中跳过 MouseMove；不 ack，留待外层（宿主可重推或下轮再取）。
			// 已取出则必须 ack，否则 waitAck 挂死——改：ack ok 空结果跳过处理。
			hostpull.AckEvent(id, true, "", "")
			continue
		}
		var args map[string]interface{}
		if argsJSON != "" && argsJSON != "null" {
			_ = json.Unmarshal([]byte(argsJSON), &args)
		}
		result, err := dispatchInbox(name, args)
		if err != nil {
			hostpull.AckEvent(id, false, "", err.Error())
		} else {
			hostpull.AckEvent(id, true, result, "")
		}
		handled = true
	}
	return handled
}

func isActionMouseMoveJSON(argsJSON string) bool {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return false
	}
	t, _ := args["type"].(string)
	return t == "MouseMove" || t == "mousemove"
}

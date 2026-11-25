//go:build witbridge

package hqeda

import (
	"encoding/json"
	"strings"

	hostpull "github.com/Huaqiu-Electronics/hqpcb-go-sdk/internal/witgen/hqpcb/host/host-pull"
)

// inboxFileBridgeEnabled：witbridge 桌面轨只用 host-pull，不再扫 inbox 文件。
func inboxFileBridgeEnabled() bool { return false }

// deferredMouseMove：嵌套 host call 期间跳过的最新 MouseMove（按动作名）。
// 外层泵送时补派发，避免橡皮筋在 ApplyMeasurePrev RPC 期间丢跟手。
var deferredMouseMove = map[string]deferredMove{}

type deferredMove struct {
	name     string
	argsJSON string
}

// pumpWitPullOnce 经 host-pull.take-event 取异步 Host→Guest 事件。
//
// 有事件时分派并 ack-event，返回 true；队列空返回 false（调用方休眠再 poll）。
// 同动作 MouseMove 只保留最新一条，避免 ApplyMeasurePrev 同步 RPC 叠爆 UI。
func pumpWitPullOnce() bool {
	handled := false
	if hostCallDepth == 0 {
		handled = flushDeferredMouseMoves() || handled
	}

	type item struct {
		id, name, argsJSON string
	}
	var batch []item
	for {
		opt := hostpull.TakeEvent()
		if opt.None() {
			break
		}
		tup := opt.Value()
		batch = append(batch, item{id: tup[0], name: tup[1], argsJSON: tup[2]})
	}
	if len(batch) == 0 {
		return handled
	}

	latestMove := map[string]item{}
	var staleMoveIDs []string
	var ordered []item
	for _, it := range batch {
		if !isActionMouseMoveJSON(it.argsJSON) {
			ordered = append(ordered, it)
			continue
		}
		key := actionNameFromArgsJSON(it.argsJSON)
		if hostCallDepth > 0 {
			// 嵌套：ack 掉避免队列堵死；记下最新坐标，外层再画。
			deferredMouseMove[key] = deferredMove{name: it.name, argsJSON: it.argsJSON}
			hostpull.AckEvent(it.id, true, "", "")
			continue
		}
		if prev, ok := latestMove[key]; ok {
			staleMoveIDs = append(staleMoveIDs, prev.id)
		}
		latestMove[key] = it
	}
	for _, id := range staleMoveIDs {
		hostpull.AckEvent(id, true, "", "")
	}
	for _, it := range latestMove {
		ordered = append(ordered, it)
	}

	for _, it := range ordered {
		var args map[string]interface{}
		if it.argsJSON != "" && it.argsJSON != "null" {
			_ = json.Unmarshal([]byte(it.argsJSON), &args)
		}
		result, err := dispatchInbox(it.name, args)
		if err != nil {
			hostpull.AckEvent(it.id, false, "", err.Error())
		} else {
			hostpull.AckEvent(it.id, true, result, "")
		}
		handled = true
	}
	return handled
}

func flushDeferredMouseMoves() bool {
	if len(deferredMouseMove) == 0 {
		return false
	}
	handled := false
	for key, d := range deferredMouseMove {
		delete(deferredMouseMove, key)
		var args map[string]interface{}
		if d.argsJSON != "" && d.argsJSON != "null" {
			_ = json.Unmarshal([]byte(d.argsJSON), &args)
		}
		_, _ = dispatchInbox(d.name, args)
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
	return strings.EqualFold(strings.TrimSpace(t), "MouseMove")
}

func actionNameFromArgsJSON(argsJSON string) string {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return ""
	}
	name, _ := args["name"].(string)
	return strings.TrimSpace(name)
}

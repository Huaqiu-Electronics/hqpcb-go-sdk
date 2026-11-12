package pcb

import (
	"strconv"
	"time"
)

// GuideSpotlight 显示聚光灯挖洞与气泡（立即返回）。
//
// target 为稳定 target id（可空表示仅气泡）；text 为说明；allowSkip 控制「跳过本步」。
func GuideSpotlight(target, text string, allowSkip bool) (string, error) {
	skip := "1"
	if !allowSkip {
		skip = "0"
	}
	return CmdletLine("guide.spotlight",
		KvToken("target", target),
		KvToken("text", text),
		KvToken("allow_skip", skip))
}

// GuideWaitClick 阻塞直到挖洞被点、跳过、退出或超时。
//
// 返回宿主结果词：ok / skip / exit / timeout / error:...
// timeoutMs≤0 时由宿主使用默认上限（约 30 分钟）。
func GuideWaitClick(target, text string, allowSkip bool, timeoutMs int) (string, error) {
	skip := "1"
	if !allowSkip {
		skip = "0"
	}
	parts := []string{
		KvToken("target", target),
		KvToken("text", text),
		KvToken("allow_skip", skip),
	}
	if timeoutMs > 0 {
		parts = append(parts, KvToken("timeout_ms", strconv.Itoa(timeoutMs)))
	}
	line := Line("guide.wait.click", parts...)
	return callFor("cmdlet", map[string]interface{}{"line": line}, 30*time.Minute)
}

// GuideWaitCmdlet 阻塞直到指定 cmdlet 名成功执行一次。
func GuideWaitCmdlet(name string, timeoutMs int) (string, error) {
	parts := []string{KvToken("name", name)}
	if timeoutMs > 0 {
		parts = append(parts, KvToken("timeout_ms", strconv.Itoa(timeoutMs)))
	}
	line := Line("guide.wait.cmdlet", parts...)
	return callFor("cmdlet", map[string]interface{}{"line": line}, 30*time.Minute)
}

// GuideBubble 显示无挖洞居中气泡。
func GuideBubble(text string) (string, error) {
	return CmdletLine("guide.bubble", KvToken("text", text))
}

// GuideMarkDone 标记向导完成。
func GuideMarkDone(id string) (string, error) {
	return CmdletLine("guide.mark_done", KvToken("id", id))
}

// GuideStop 请求宿主停止当前向导。
func GuideStop() (string, error) {
	return Cmdlet("guide.stop")
}

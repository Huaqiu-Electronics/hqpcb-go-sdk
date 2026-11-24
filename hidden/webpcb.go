package hidden

import (
	"encoding/json"
	"strings"

	hqeda "github.com/Huaqiu-Electronics/hqpcb-go-sdk"
)

// EnsureWebPcb 拉起旁路 hqpcb-web（Go HTTP+SSE），返回 baseUrl / viewUrl。
func EnsureWebPcb(boardPath string) (baseURL, viewURL string, err error) {
	raw, err := hqeda.HostOp("ensure_webpcb", map[string]interface{}{
		"board": strings.TrimSpace(boardPath),
	})
	if err != nil {
		return "", "", err
	}
	var out struct {
		URL     string `json:"url"`
		ViewURL string `json:"viewUrl"`
	}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return raw, raw, nil
	}
	return out.URL, out.ViewURL, nil
}

// StopWebPcb 关闭本机 hqpcb-web。
func StopWebPcb() error {
	_, err := hqeda.HostOp("stop_webpcb", map[string]interface{}{})
	return err
}

// OpenURL 用系统默认浏览器打开 URL。
func OpenURL(url string) error {
	_, err := hqeda.HostOp("open_url", map[string]interface{}{
		"url": strings.TrimSpace(url),
	})
	return err
}

// WebPcbNotify 将补丁摘要 POST 到 hqpcb-web /api/pcb/notify（宿主代发）。
func WebPcbNotify(baseURL string, payload map[string]interface{}) error {
	if payload == nil {
		payload = map[string]interface{}{}
	}
	_, err := hqeda.HostOp("webpcb_notify", map[string]interface{}{
		"baseUrl": strings.TrimSpace(baseURL),
		"payload": payload,
	})
	return err
}

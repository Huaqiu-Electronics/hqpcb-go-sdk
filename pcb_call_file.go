//go:build !witbridge

package hqeda

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
	hostCallDepth++
	defer func() { hostCallDepth-- }()
	deadline := time.Now().Add(maxWait)
	for time.Now().Before(deadline) {
		// 常驻插件：等待宿主 resp 时仍泵送 inbox，避免 command_push→command.do 嵌套死锁。
		if PluginMode() {
			pumpInboxOnce()
		}
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

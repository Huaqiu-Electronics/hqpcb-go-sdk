//go:build !witbridge

package hqeda

import "time"

// call 发送一次 host 请求并等待响应（轮询，最长约 120s）。
func call(op string, args map[string]interface{}) (string, error) {
	return callFor(op, args, 120*time.Second)
}

// callFor 同 call，但可指定最长等待时间（向导 wait.* 用更长超时）。
func callFor(op string, args map[string]interface{}, maxWait time.Duration) (string, error) {
	return callFileFor(op, args, maxWait)
}

//go:build witbridge

// WIT 双轨：Guest 经 hqpcb:host/op import 调宿主（需 wasip2 + pcb-wit-run / wac）。
// 构建：tinygo build -tags witbridge -target=wasip2-hqpcb.json …
//
// 全部 Guest→Host 走同步 WIT；嵌套 Host→Guest 经 RPC kind=guest → guest-events。
package hqeda

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/Huaqiu-Electronics/hqpcb-go-sdk/internal/witgen/hqpcb/host/op"
	"go.bytecodealliance.org/cm"
)

func call(opName string, args map[string]interface{}) (string, error) {
	return callFor(opName, args, 120*time.Second)
}

func callFor(opName string, args map[string]interface{}, maxWait time.Duration) (string, error) {
	_ = maxWait // sync WIT；超时由宿主侧负责
	return callWitSync(opName, args)
}

func callWitSync(opName string, args map[string]interface{}) (string, error) {
	hostCallDepth++
	defer func() { hostCallDepth-- }()

	switch opName {
	case "log":
		msg, _ := args["message"].(string)
		op.Log(msg)
		return "", nil
	case "cmdlet":
		line, _ := args["line"].(string)
		return resultString(op.Cmdlet(line))
	case "query":
		sql, _ := args["sql"].(string)
		return resultString(op.Query(sql))
	case "doAction":
		method, _ := args["method"].(string)
		body, _ := args["body"].(string)
		return resultString(op.DoAction(method, body))
	default:
		raw, err := json.Marshal(args)
		if err != nil {
			return "", err
		}
		return resultString(op.Call(opName, string(raw)))
	}
}

func resultString(r cm.Result[string, string, string]) (string, error) {
	ok, err, isErr := r.Result()
	if isErr {
		if err == "" {
			return "", errors.New("pcb: host call failed")
		}
		return "", errors.New(err)
	}
	return ok, nil
}

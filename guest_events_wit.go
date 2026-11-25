//go:build witbridge

package hqeda

import (
	"encoding/json"

	guestevents "github.com/Huaqiu-Electronics/hqpcb-go-sdk/internal/witgen/hqpcb/host/guest-events"
	"go.bytecodealliance.org/cm"
)

func init() {
	guestevents.Exports.PostEvent = postEventWit
}

// postEventWit 实现 hqpcb:host/guest-events#post-event（与 inbox 同派发）。
func postEventWit(name string, argsJSON string) cm.Result[string, string, string] {
	var args map[string]interface{}
	if s := argsJSON; s != "" && s != "null" {
		if err := json.Unmarshal([]byte(s), &args); err != nil {
			return cm.Err[cm.Result[string, string, string]]("invalid args-json: " + err.Error())
		}
	}
	result, err := dispatchInbox(name, args)
	if err != nil {
		return cm.Err[cm.Result[string, string, string]](err.Error())
	}
	return cm.OK[cm.Result[string, string, string]](result)
}

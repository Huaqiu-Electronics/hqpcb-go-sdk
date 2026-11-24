package hqeda

// HostOp 是不稳定的宿主桥逃生舱，仅供一等内置插件（见 package hidden）。
// 不构成公开 Guest SDK 契约，可无通知变更。
func HostOp(op string, args map[string]interface{}) (string, error) {
	if args == nil {
		args = map[string]interface{}{}
	}
	return call(op, args)
}

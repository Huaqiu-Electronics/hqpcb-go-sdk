//go:build !witbridge

package hqeda

// inboxFileBridgeEnabled：非 witbridge 仍走 inbox/inbox_ack 文件桥。
func inboxFileBridgeEnabled() bool { return true }

// pumpWitPullOnce：非 witbridge 无 host-pull。
func pumpWitPullOnce() bool { return false }

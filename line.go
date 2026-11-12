package pcb

import (
	"fmt"
	"strings"
)

// EncodeValue 将明文编成 cmdlet 键值右侧百分号串。
//
// 对齐宿主 CmdletKvArgs / QUrl::toPercentEncoding：仅保留
// ALPHA / DIGIT / "-" / "." / "_" / "~"，其余按 UTF-8 字节写成 %XX。
// 空格编码为 %20（不用 +），以便 fromPercentEncoding 还原。
func EncodeValue(raw string) string {
	if raw == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(raw) * 3)
	for i := 0; i < len(raw); i++ {
		c := raw[i]
		if isUnreserved(c) {
			b.WriteByte(c)
			continue
		}
		fmt.Fprintf(&b, "%%%02X", c)
	}
	return b.String()
}

// isUnreserved 判定 RFC 3986 unreserved，与 Qt toPercentEncoding 默认保留集一致。
func isUnreserved(c byte) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') ||
		c == '-' || c == '.' || c == '_' || c == '~'
}

// KvToken 拼一条 key=enc 令牌；key 空时返回空串。
func KvToken(key, value string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return ""
	}
	return key + "=" + EncodeValue(value)
}

// Line 拼完整 cmdlet 命令行：name + 空白分隔参数。
//
// 键值参数请先经 KvToken；位置参数直接传入明文（勿含空白，宿主按空白切词）。
func Line(name string, parts ...string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	out := make([]string, 0, 1+len(parts))
	out = append(out, name)
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return strings.Join(out, " ")
}

// CmdletLine 执行 Line 拼出的命令行，语义同 Cmdlet。
func CmdletLine(name string, parts ...string) (string, error) {
	return Cmdlet(Line(name, parts...))
}

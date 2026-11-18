package hqeda

import (
	"strings"
)

// DesignWriter 最小设计写面（FlightIPC / Board / FlatRpc 适配）。
type DesignWriter interface {
	DesignLease(token string) (string, error)
	DesignSet(key, value, token string) (string, error)
	DesignGet(key string) (string, error)
	DesignRelease(token string) (string, error)
	DesignMeta() (string, error)
}

// ConfigWriter 最小配置写面。
type ConfigWriter interface {
	ConfigSet(key, value string) (string, error)
	ConfigGet(key string) (string, error)
}

// WriteProps lease → 批量 set → release；返回读回 map。
//
// token 空时用 "sdk-write"。调用方无需单独拿租约。
func WriteProps(api DesignWriter, props map[string]string, token string) (map[string]string, error) {
	if strings.TrimSpace(token) == "" {
		token = "sdk-write"
	}
	if _, err := api.DesignLease(token); err != nil {
		return nil, err
	}
	defer func() { _, _ = api.DesignRelease(token) }()

	for k, v := range props {
		if _, err := api.DesignSet(k, v, token); err != nil {
			return nil, err
		}
	}
	out := make(map[string]string, len(props))
	for k := range props {
		v, err := api.DesignGet(k)
		if err != nil {
			return nil, err
		}
		out[k] = v
	}
	return out, nil
}

// WriteConfig 写配置并读回。
func WriteConfig(api ConfigWriter, key, value string) (string, error) {
	if _, err := api.ConfigSet(key, value); err != nil {
		return "", err
	}
	return api.ConfigGet(key)
}

// ParseWsField 从行协议取 field= 值。
func ParseWsField(snapshot, field string) string {
	prefix := field + "="
	for _, line := range strings.Split(snapshot, "\n") {
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, prefix))
		}
	}
	return ""
}

// WorkspaceID 优先 id=，否则 current=（list/open 快照）。
func WorkspaceID(snapshot string) string {
	if id := ParseWsField(snapshot, "id"); id != "" {
		return id
	}
	return ParseWsField(snapshot, "current")
}

package pcb

// ScriptGoCompile 编译 TinyGo 脚本为 WASI .wasm。
//
// path 为 .go 文件或目录；force 为 true 时附加 force=1。
// 参数走 path= 百分号编码，可含空格路径。
func ScriptGoCompile(path string, force bool) (string, error) {
	parts := []string{KvToken("path", path)}
	if force {
		parts = append(parts, KvToken("force", "1"))
	}
	return CmdletLine("script.go.compile", parts...)
}

// ScriptGoRun 编译（如需）并运行 TinyGo / .wasm 脚本。
//
// 参数走 path= 百分号编码，可含空格路径。
func ScriptGoRun(path string) (string, error) {
	return CmdletLine("script.go.run", KvToken("path", path))
}

// ScriptGoInfo 查询宿主 TinyGo / Wasmtime / Guest SDK 诊断信息。
func ScriptGoInfo() (string, error) {
	return Cmdlet("script.go.info")
}

// ScriptGoCacheClear 清空 GoScript 编译缓存。
func ScriptGoCacheClear() (string, error) {
	return Cmdlet("script.go.cache.clear")
}

// ScriptRun 运行 .cmdlet 文本脚本（非 Go/WASM）。
func ScriptRun(path string) (string, error) {
	if path == "" {
		return Cmdlet("script.run")
	}
	return CmdletLine("script.run", path)
}

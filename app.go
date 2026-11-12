package pcb

// Plugins 列出已装载插件（对齐 cmdlet plugins）。
func Plugins() (string, error) {
	return Cmdlet("plugins")
}

// Help 查询 cmdlet 帮助；name 空则列出总览。
func Help(name string) (string, error) {
	if name == "" {
		return Cmdlet("help")
	}
	return CmdletLine("help", name)
}

// WindowList 列出主窗口（对齐 cmdlet window.list）。
func WindowList() (string, error) {
	return Cmdlet("window.list")
}

// WindowActivate 按窗口序号或 id 激活主窗。
func WindowActivate(id string) (string, error) {
	return CmdletLine("window.activate", id)
}

// WindowClose 关闭指定主窗。
func WindowClose(id string) (string, error) {
	return CmdletLine("window.close", id)
}

// LoadPlugin 装载插件（路径为 plugin.yaml 或插件目录；勿含空白）。
func LoadPlugin(path string) (string, error) {
	return CmdletLine("loadplugin", path)
}

// UnloadPlugin 卸载插件（按 id 或路径，随宿主约定）。
func UnloadPlugin(idOrPath string) (string, error) {
	return CmdletLine("unloadplugin", idOrPath)
}

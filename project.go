package pcb

// ProjectOpenOpts 控制 project.open 的可选开关。
type ProjectOpenOpts struct {
	ReadOnly      bool // 只读打开（readonly=1）
	NoSaveAs      bool // 禁止另存（no-save-as=1）；通常配合 ReadOnly
	NoRecent      bool // 不写入最近列表（no-recent=1）
	SilentRestore bool // AI 静默恢复备份（silent-restore=1）
}

// ProjectOpen 打开工程；path 空则由宿主弹对话框。
//
// 参数走 path=/readonly= 百分号编码契约（CmdletKvArgs），可含空格路径。
func ProjectOpen(path string, opts ProjectOpenOpts) (string, error) {
	parts := make([]string, 0, 5)
	if path != "" {
		parts = append(parts, KvToken("path", path))
	}
	if opts.ReadOnly {
		parts = append(parts, KvToken("readonly", "1"))
	}
	if opts.NoSaveAs {
		parts = append(parts, KvToken("no-save-as", "1"))
	}
	if opts.NoRecent {
		parts = append(parts, KvToken("no-recent", "1"))
	}
	if opts.SilentRestore {
		parts = append(parts, KvToken("silent-restore", "1"))
	}
	return CmdletLine("project.open", parts...)
}

// ProjectSave 保存当前工程。
func ProjectSave() (string, error) {
	return Cmdlet("project.save")
}

// ProjectSaveAs 另存为；path 为空则由宿主弹对话框。
//
// 位置参数：路径勿含空白（宿主空白切词无引号）。
func ProjectSaveAs(path string) (string, error) {
	if path == "" {
		return Cmdlet("project.save_as")
	}
	return CmdletLine("project.save_as", path)
}

// ProjectNew 创建新工程（对齐 cmdlet project.new；具体参数由宿主默认）。
func ProjectNew() (string, error) {
	return Cmdlet("project.new")
}

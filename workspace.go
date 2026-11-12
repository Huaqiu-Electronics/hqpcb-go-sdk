package pcb

// WorkspaceList 列出运行时工作区（对齐 cmdlet workspace.list）。
func WorkspaceList() (string, error) {
	return Cmdlet("workspace.list")
}

// WorkspaceNew 创建新工作区（对齐 cmdlet workspace.new）。
func WorkspaceNew() (string, error) {
	return Cmdlet("workspace.new")
}

// WorkspaceActivate 按 objectName 或 stable id 激活工作区。
//
// id 勿含空白（宿主按空白切词，无引号支持）。
func WorkspaceActivate(id string) (string, error) {
	return CmdletLine("workspace.activate", id)
}

// WorkspaceClose 按 id 关闭工作区。
func WorkspaceClose(id string) (string, error) {
	return CmdletLine("workspace.close", id)
}

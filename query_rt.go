package pcb

// 常用 runtime.* 虚拟表查询封装。
//
// SQL 列与宿主登记表对齐（见 AGENTS.md §5 / LeftPanelDemo）；
// 只读封装，写回仍走 cmdlet / host method。

// SQL 常量：plugins 列表（默认列）。
const sqlRuntimePlugins = "select id, name, version, state from runtime.plugins order by id"

// SQL 常量：层显示 schema（与 LeftPanelDemo SQL_LAYER_SCHEMA_LAYERS 对齐）。
const sqlLayerSchemaLayers = "" +
	"select top_layer_type, top_layer_id, top_layer_title, ordinal, " +
	"layer_type, layer_id, title, visible, selectable, color, alpha, " +
	"fill_method, fill_angle, fill_pitch, fill_width, is_current " +
	"from runtime.layer_schema_layers " +
	"order by top_layer_type, ordinal"

// SQL 常量：视图选项。
const sqlViewOptions = "select * from runtime.view_options"

// SQL 常量：cmdlet 历史。
const sqlCmdletHistory = "select * from runtime.cmdlet_history"

// SQL 常量：命令历史。
const sqlCommandHistory = "select * from runtime.command_history"

// SQL 常量：配置中心项。
const sqlConfigItems = "select * from runtime.config_items"

// QueryPlugins 查询 runtime.plugins；limit<=0 表示不限制行数。
func QueryPlugins(limit int) (QueryTable, error) {
	return QueryTableResult(sqlLimit(sqlRuntimePlugins, limit))
}

// QueryLayers 查询 runtime.layer_schema_layers（层可见性/颜色/填充等显示配置）。
func QueryLayers() (QueryTable, error) {
	return QueryTableResult(sqlLayerSchemaLayers)
}

// QueryViewOptions 查询 runtime.view_options。
func QueryViewOptions() (QueryTable, error) {
	return QueryTableResult(sqlViewOptions)
}

// QueryCmdletHist 查询 runtime.cmdlet_history；limit<=0 不限制。
func QueryCmdletHist(limit int) (QueryTable, error) {
	return QueryTableResult(sqlLimit(sqlCmdletHistory, limit))
}

// QueryCmdHist 查询 runtime.command_history；limit<=0 不限制。
func QueryCmdHist(limit int) (QueryTable, error) {
	return QueryTableResult(sqlLimit(sqlCommandHistory, limit))
}

// QueryConfigItems 查询 runtime.config_items。
func QueryConfigItems() (QueryTable, error) {
	return QueryTableResult(sqlConfigItems)
}

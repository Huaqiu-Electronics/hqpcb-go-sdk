package pcb

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// QueryTable 是宿主 Query 返回 JSON 的结构化形态。
//
// 字段与 telnet / runtime.query / GoScript 桥同源：columns 为列名，
// rows 为字符串单元格矩阵。错误经 call 的 error 返回，不嵌在本 JSON 内。
type QueryTable struct {
	Columns []string   `json:"columns"` // 列名，顺序与每行单元格对齐。
	Rows    [][]string `json:"rows"`    // 行数据；允许为空表。
}

// QueryTableResult 执行 SQL 并解析为 QueryTable。
func QueryTableResult(sql string) (QueryTable, error) {
	raw, err := Query(sql)
	if err != nil {
		return QueryTable{}, err
	}
	return ParseQueryTable(raw)
}

// QueryMaps 执行 SQL，返回「列名→单元格」行映射切片（便捷封装）。
func QueryMaps(sql string) ([]map[string]string, error) {
	tab, err := QueryTableResult(sql)
	if err != nil {
		return nil, err
	}
	return tab.Maps(), nil
}

// ParseQueryTable 将 Query 返回的 JSON 文本解析为 QueryTable。
func ParseQueryTable(raw string) (QueryTable, error) {
	var t QueryTable
	if err := json.Unmarshal([]byte(raw), &t); err != nil {
		return QueryTable{}, fmt.Errorf("pcb: parse query json: %w", err)
	}
	if t.Columns == nil {
		t.Columns = []string{}
	}
	if t.Rows == nil {
		t.Rows = [][]string{}
	}
	return t, nil
}

// Len 返回行数。
func (t QueryTable) Len() int {
	return len(t.Rows)
}

// Empty 在无行时为 true（列定义可仍非空）。
func (t QueryTable) Empty() bool {
	return len(t.Rows) == 0
}

// ColIndex 返回列名下标；不存在时为 -1。
func (t QueryTable) ColIndex(col string) int {
	for i, name := range t.Columns {
		if name == col {
			return i
		}
	}
	return -1
}

// HasColumn 判断列名是否存在。
func (t QueryTable) HasColumn(col string) bool {
	return t.ColIndex(col) >= 0
}

// Row 取第 row 行原始单元格切片；越界时返回 nil 与 false。
func (t QueryTable) Row(row int) ([]string, bool) {
	if row < 0 || row >= len(t.Rows) {
		return nil, false
	}
	return t.Rows[row], true
}

// Cell 取第 row 行、列名为 col 的单元格；越界或无列时返回空串与 false。
func (t QueryTable) Cell(row int, col string) (string, bool) {
	cells, ok := t.Row(row)
	if !ok {
		return "", false
	}
	idx := t.ColIndex(col)
	if idx < 0 || idx >= len(cells) {
		return "", false
	}
	return cells[idx], true
}

// RowMap 将第 row 行转为列名→值映射；越界时返回 nil 与 false。
//
// 行内缺列时对应值为空串；多余单元格忽略。
func (t QueryTable) RowMap(row int) (map[string]string, bool) {
	cells, ok := t.Row(row)
	if !ok {
		return nil, false
	}
	m := make(map[string]string, len(t.Columns))
	for i, name := range t.Columns {
		if i < len(cells) {
			m[name] = cells[i]
		} else {
			m[name] = ""
		}
	}
	return m, true
}

// Maps 将全部行转为列名→值映射切片（无行时返回空切片非 nil）。
func (t QueryTable) Maps() []map[string]string {
	out := make([]map[string]string, 0, len(t.Rows))
	for i := range t.Rows {
		m, ok := t.RowMap(i)
		if ok {
			out = append(out, m)
		}
	}
	return out
}

// sqlLimit 在 base SQL 后附加 LIMIT；limit<=0 表示不加限制。
func sqlLimit(base string, limit int) string {
	if limit <= 0 {
		return base
	}
	return base + " limit " + strconv.Itoa(limit)
}

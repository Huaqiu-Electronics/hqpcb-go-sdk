package pcb

import "testing"

func TestParseQueryTable(t *testing.T) {
	raw := `{"columns":["id","name"],"rows":[["builtin.pcb.host","PcbHost"]]}`
	tab, err := ParseQueryTable(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(tab.Columns) != 2 || tab.Len() != 1 {
		t.Fatalf("unexpected shape: %+v", tab)
	}
	v, ok := tab.Cell(0, "name")
	if !ok || v != "PcbHost" {
		t.Fatalf("Cell name=%q ok=%v", v, ok)
	}
	if _, ok := tab.Cell(0, "missing"); ok {
		t.Fatal("missing column should fail")
	}
}

func TestQueryTableMapsAndRow(t *testing.T) {
	tab, err := ParseQueryTable(`{"columns":["a","b"],"rows":[["1","2"],["3","4"]]}`)
	if err != nil {
		t.Fatal(err)
	}
	if tab.Empty() || !tab.HasColumn("a") || tab.ColIndex("b") != 1 {
		t.Fatalf("meta: empty=%v hasA=%v idxB=%d", tab.Empty(), tab.HasColumn("a"), tab.ColIndex("b"))
	}
	row, ok := tab.Row(1)
	if !ok || len(row) != 2 || row[0] != "3" {
		t.Fatalf("Row: %v ok=%v", row, ok)
	}
	m, ok := tab.RowMap(0)
	if !ok || m["a"] != "1" || m["b"] != "2" {
		t.Fatalf("RowMap: %v ok=%v", m, ok)
	}
	maps := tab.Maps()
	if len(maps) != 2 || maps[1]["b"] != "4" {
		t.Fatalf("Maps: %v", maps)
	}
}

func TestParseQueryTableNilFields(t *testing.T) {
	tab, err := ParseQueryTable(`{}`)
	if err != nil {
		t.Fatal(err)
	}
	if tab.Columns == nil || tab.Rows == nil || !tab.Empty() {
		t.Fatalf("nil fields should normalize: %+v", tab)
	}
}

func TestSqlLimit(t *testing.T) {
	if got := sqlLimit("select 1", 0); got != "select 1" {
		t.Fatalf("limit 0: %q", got)
	}
	if got := sqlLimit("select 1", 5); got != "select 1 limit 5" {
		t.Fatalf("limit 5: %q", got)
	}
}

package pcb

import "testing"

func TestEncodeValue(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"", ""},
		{"abc", "abc"},
		{"a b", "a%20b"},
		{"C:/HQ PCB/x.hqpcb", "C%3A%2FHQ%20PCB%2Fx.hqpcb"},
		{"中", "%E4%B8%AD"},
	}
	for _, c := range cases {
		if got := EncodeValue(c.in); got != c.want {
			t.Fatalf("EncodeValue(%q)=%q want %q", c.in, got, c.want)
		}
	}
}

func TestProjectOpenLine(t *testing.T) {
	line := Line("project.open", KvToken("path", "/tmp/a b.hqpcb"), KvToken("readonly", "1"))
	if line != "project.open path=%2Ftmp%2Fa%20b.hqpcb readonly=1" {
		t.Fatalf("unexpected line: %q", line)
	}
}

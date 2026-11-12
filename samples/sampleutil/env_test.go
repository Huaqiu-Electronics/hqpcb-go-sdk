package sampleutil

import "testing"

func TestResolvePipe(t *testing.T) {
	cases := []struct {
		argv []string
		pipe string
		gui  bool
		rest []string
	}{
		{[]string{"x"}, "", false, nil},
		{[]string{"x", "gui"}, "", true, nil},
		{[]string{"x", "--gui"}, "", true, nil},
		{[]string{"x", "hqpcb-gui"}, "hqpcb-gui", true, nil},
		{[]string{"x", "hqpcb-svc", "/b"}, "hqpcb-svc", false, []string{"/b"}},
		{[]string{"x", "--endpoint=gui", "a"}, "", true, []string{"a"}},
	}
	for _, c := range cases {
		p, g, r := ResolvePipe(c.argv)
		if p != c.pipe || g != c.gui {
			t.Fatalf("argv=%v got pipe=%q gui=%v want %q %v", c.argv, p, g, c.pipe, c.gui)
		}
		if len(r) != len(c.rest) {
			t.Fatalf("rest %v vs %v", r, c.rest)
		}
		for i := range r {
			if r[i] != c.rest[i] {
				t.Fatalf("rest %v vs %v", r, c.rest)
			}
		}
	}
}

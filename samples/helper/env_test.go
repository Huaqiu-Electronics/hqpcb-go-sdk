package helper

import (
	"os"
	"testing"
)

func TestResolvePipe(t *testing.T) {
	_ = os.Unsetenv("PCB_IPC_ENDPOINT")
	cases := []struct {
		argv []string
		gui  bool
		rest int
	}{
		{[]string{"p", "headless", "a.hqpcb"}, false, 1},
		{[]string{"p", "gui"}, true, 0},
		{[]string{"p", "hqpcb-svc"}, false, 0},
	}
	for _, c := range cases {
		p, g, r := ResolvePipe(c.argv)
		_ = p
		if g != c.gui || len(r) != c.rest {
			t.Fatalf("argv=%v gui=%v rest=%v", c.argv, g, r)
		}
	}
}

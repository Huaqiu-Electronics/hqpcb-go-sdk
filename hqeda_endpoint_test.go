package hqeda

import (
	"os"
	"testing"
)

func TestResolveEndpoint(t *testing.T) {
	_ = os.Unsetenv("PCB_IPC_ENDPOINT")
	ep, err := ResolveEndpoint([]string{"prog", "headless", "/tmp/a.hqpcb"})
	if err != nil {
		t.Fatal(err)
	}
	if ep.Gui || ep.Pipe != "" || len(ep.Rest) != 1 || ep.Rest[0] != "/tmp/a.hqpcb" {
		t.Fatalf("got %+v", ep)
	}
	ep, err = ResolveEndpoint([]string{"prog", "gui"})
	if err != nil {
		t.Fatal(err)
	}
	if !ep.Gui {
		t.Fatalf("expected gui")
	}
	_, err = ResolveEndpoint([]string{"prog", "grpc://x"})
	if err == nil {
		t.Fatal("expected error for URI")
	}
}

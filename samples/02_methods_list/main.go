package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/Huaqiu-Electronics/hqpcb-go-sdk"
	"github.com/Huaqiu-Electronics/hqpcb-go-sdk/samples/sampleutil"
)

func main() {
	pipe, gui, _ := sampleutil.ResolvePipe(os.Args)
	cli := pcb.NewFlightIPC()
	if err := cli.Connect(pipe, gui); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer cli.Disconnect()
	catalog, err := cli.MethodsList()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("# pipe=%s role=%s interactive=%v\n", cli.Pipe(), cli.PeerRole(), cli.PeerCaps().Interactive)
	fmt.Println(strings.TrimRight(catalog, "\n"))
	core, ix := 0, 0
	for _, ln := range strings.Split(catalog, "\n") {
		if strings.Contains(ln, "surface=core") {
			core++
		}
		if strings.Contains(ln, "surface=interactive") {
			ix++
		}
	}
	fmt.Printf("# count core=%d interactive=%d\n", core, ix)
}

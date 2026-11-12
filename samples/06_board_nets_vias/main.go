package main

import (
	"fmt"
	"os"

	"github.com/Huaqiu-Electronics/hqpcb-go-sdk"
	"github.com/Huaqiu-Electronics/hqpcb-go-sdk/samples/sampleutil"
)

func main() {
	pipe, gui, rest := sampleutil.ResolvePipe(os.Args)
	arg := ""
	if len(rest) > 0 {
		arg = rest[0]
	}
	path, err := sampleutil.BoardPath(arg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	cli := pcb.NewFlightIPC()
	if err := cli.Connect(pipe, gui); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer cli.Disconnect()
	sampleutil.RequireWsDesign(cli.Pipe(), cli.PeerRole(), cli.PeerCaps().Workspace, cli.PeerCaps().Design)
	fmt.Printf("pipe=%s\n", cli.Pipe())
	board, err := pcb.OpenBoard(pcb.BindFlightIPC(cli), path, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("path=%s\n", board.Path)
	nets, err := board.Nets()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("net_count=%d\n", len(nets))
	for i, n := range nets {
		if i >= 10 {
			break
		}
		fmt.Printf("  net code=%v name=%q class=%q\n", n["code"], n["name"], n["net_class"])
	}
	vias, err := board.Vias()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("via_count=%d\n", len(vias))
	for i, v := range vias {
		if i >= 5 {
			break
		}
		fmt.Printf("  via id=%v net=%q xy=(%v,%v) layers=%v->%v\n",
			v["id"], v["net"], v["x_mm"], v["y_mm"], v["layer_from_id"], v["layer_to_id"])
	}
}

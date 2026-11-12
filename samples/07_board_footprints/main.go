package main

import (
	"fmt"
	"os"

	"github.com/Huaqiu-Electronics/hqpcb-go-sdk"
	"github.com/Huaqiu-Electronics/hqpcb-go-sdk/samples/sampleutil"
)

func truthy(v any) bool {
	switch x := v.(type) {
	case bool:
		return x
	case float64:
		return x != 0
	case int:
		return x != 0
	case string:
		return x == "1" || x == "true"
	default:
		return false
	}
}

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
	fps, err := board.Footprints()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("footprint_count=%d\n", len(fps))
	for i, fp := range fps {
		if i >= 8 {
			break
		}
		fmt.Printf("  %v pkg=%q xy=(%v,%v) placed=%v\n",
			fp["refdes"], fp["package"], fp["x_mm"], fp["y_mm"], fp["placed"])
	}
	pads, err := board.Pads()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("pad_count=%d\n", len(pads))
	for i, p := range pads {
		if i >= 5 {
			break
		}
		fmt.Printf("  pin=%q/%q net_id=%v xy=(%v,%v)\n",
			p["pin_number"], p["pin_name"], p["net_id"], p["x_mm"], p["y_mm"])
	}
	layers, err := board.Layers()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	copper := 0
	for _, ly := range layers {
		if truthy(ly["is_conductor"]) {
			copper++
		}
	}
	fmt.Printf("layer_count=%d\n", len(layers))
	fmt.Printf("copper_layer_count=%d\n", copper)
	for i, ly := range layers {
		if i >= 6 {
			break
		}
		fmt.Printf("  id=%v name=%q conductor=%v\n", ly["layer_id"], ly["name"], ly["is_conductor"])
	}
}

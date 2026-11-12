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
	zones, err := board.Zones()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("zone_count=%d\n", len(zones))
	for i, z := range zones {
		if i >= 8 {
			break
		}
		fmt.Printf("  id=%v name=%q net=%q filled=%v bbox=(%v,%v)-(%v,%v)\n",
			z["id"], z["name"], z["net"], z["is_filled"],
			z["min_x_mm"], z["min_y_mm"], z["max_x_mm"], z["max_y_mm"])
	}
	drawings, err := board.Drawings()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("drawing_count=%d\n", len(drawings))
	for i, d := range drawings {
		if i >= 6 {
			break
		}
		fmt.Printf("  kind=%v id=%v name=%q xy=(%v,%v) filled=%v\n",
			d["kind"], d["id"], d["name"], d["x_mm"], d["y_mm"], d["is_filled"])
	}
	bbox, err := board.BoundingBox()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("bbox=(%.4f,%.4f)-(%.4f,%.4f) size=%.4fx%.4f mm\n",
		bbox.MinXMm, bbox.MinYMm, bbox.MaxXMm, bbox.MaxYMm, bbox.WidthMm(), bbox.HeightMm())
}

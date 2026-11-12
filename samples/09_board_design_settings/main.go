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
	ds, err := board.DesignSettings()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("default_track_width_mm=%s min=%s clearance=%s\n",
		ds["default_track_width_mm"], ds["min_track_width_mm"], ds["default_clearance_mm"])
	fmt.Printf("default_via=%q dia=%s drill=%s\n",
		ds["default_via"], ds["via_diameter_mm"], ds["via_drill_mm"])
	fmt.Printf("counts net_classes=%s track_widths=%s vias=%s\n",
		ds["net_class_count"], ds["track_width_rule_count"], ds["via_rule_count"])
	classes, err := board.NetClasses()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("net_class_count=%d\n", len(classes))
	for _, nc := range classes {
		fmt.Printf("  %q members=%v width=%q via=%q\n",
			nc["name"], nc["member_count"], nc["width_rule"], nc["default_via"])
	}
	widths, err := board.TrackWidths()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("track_width_count=%d\n", len(widths))
	for i, w := range widths {
		if i >= 6 {
			break
		}
		fmt.Printf("  %q layer=%v def=%v min=%v\n",
			w["name"], w["layer"], w["default_width_mm"], w["min_width_mm"])
	}
	vias, err := board.ViaDefs()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("via_def_count=%d\n", len(vias))
	for _, v := range vias {
		fmt.Printf("  %q is_via=%v dia=%v drill=%v\n",
			v["padstack_name"], v["is_via"], v["diameter_mm"], v["drill_mm"])
	}
}

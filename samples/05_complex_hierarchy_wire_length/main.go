package main

import (
	"fmt"
	"os"
	"sort"

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
	fmt.Println("--- Board.open ---")
	board, err := pcb.OpenBoard(pcb.BindFlightIPC(cli), path, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("path=%s\n", board.Path)
	gen, _ := cli.WsGen()
	fmt.Printf("gen=%d\n", gen)
	fmt.Println("--- Board.tracks (rpc=board.tracks) ---")
	tracks, err := board.Tracks(false)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("track_count=%d\n", len(tracks))
	if len(tracks) > 0 {
		t0 := tracks[0]
		fmt.Printf("sample[0] net=%s layer=%s width=%.4f length=%.4f\n",
			t0.GetNetname(), t0.Layer, t0.WidthMm, t0.Length())
	}
	fmt.Println("--- length by net (Go OO) ---")
	byNet, err := board.LengthByNet()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	type kv struct {
		k string
		v float64
	}
	items := make([]kv, 0, len(byNet))
	total := 0.0
	for k, v := range byNet {
		items = append(items, kv{k, v})
		total += v
	}
	sort.Slice(items, func(i, j int) bool { return items[i].v > items[j].v })
	fmt.Printf("nets_with_tracks=%d\n", len(byNet))
	fmt.Printf("total_track_length_mm=%.4f\n", total)
	for i, it := range items {
		if i >= 15 {
			break
		}
		segs := 0
		for _, t := range tracks {
			if t.GetNetname() == it.k {
				segs++
			}
		}
		fmt.Printf("  %s: %.4f  (segments=%d)\n", it.k, it.v, segs)
	}
}

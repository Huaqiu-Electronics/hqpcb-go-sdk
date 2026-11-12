package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Huaqiu-Electronics/hqpcb-go-sdk"
	"github.com/Huaqiu-Electronics/hqpcb-go-sdk/samples/sampleutil"
)

func main() {
	pipe, gui, rest := sampleutil.ResolvePipe(os.Args)
	cli := pcb.NewFlightIPC()
	if err := cli.Connect(pipe, gui); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer cli.Disconnect()
	sampleutil.RequireWsDesign(cli.Pipe(), cli.PeerRole(), cli.PeerCaps().Workspace, cli.PeerCaps().Design)
	caps := cli.PeerCaps()
	if !caps.Workspace || !caps.Design {
		fmt.Fprintf(os.Stderr, "peer lacks workspace/design (pipe=%s role=%s); use --role service\n",
			cli.Pipe(), cli.PeerRole())
		os.Exit(1)
	}
	var board string
	if len(rest) > 0 {
		abs, err := filepath.Abs(rest[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		board = abs
	} else {
		board = filepath.Join(os.TempDir(), "pcb_sdk_sample_board.hqpcb")
		_ = os.WriteFile(board, []byte("sdk-go-sample"), 0o644)
	}
	fmt.Printf("pipe=%s\n", cli.Pipe())
	fmt.Println("--- ws.open ---")
	snap, err := cli.WsOpen(board)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(strings.TrimRight(snap, "\n"))
	gen, err := cli.WsGen()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("gen=%d\n", gen)
	token := "go-sample-token"
	fmt.Println("--- design.lease / set / get ---")
	out, err := cli.DesignLease(token)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(strings.TrimRight(out, "\n"))
	out, err = cli.DesignSet("layer", "top", token)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(strings.TrimRight(out, "\n"))
	layer, err := cli.DesignGet("layer")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("layer=%s\n", layer)
	out, err = cli.DesignRelease(token)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(strings.TrimRight(out, "\n"))
}

package main

import (
	"fmt"
	"os"
	"strings"

	hqeda "github.com/Huaqiu-Electronics/hqpcb-go-sdk"
	"github.com/Huaqiu-Electronics/hqpcb-go-sdk/samples/helper"
)

func main() {
	_, gui, rest := helper.ResolvePipe(os.Args)
	cli, err := hqeda.NewConnection(gui)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer cli.Disconnect()
	helper.CheckEnv(cli.Pipe(), cli.PeerRole(), cli.PeerCaps().Workspace, cli.PeerCaps().Design)
	caps := cli.PeerCaps()
	if !caps.Workspace || !caps.Design {
		fmt.Fprintf(os.Stderr, "peer lacks workspace/design (pipe=%s role=%s); use --role service\n",
			cli.Pipe(), cli.PeerRole())
		os.Exit(1)
	}
	boardArg := ""
	if len(rest) > 0 {
		boardArg = rest[0]
	}
	board, err := helper.BoardPath(boardArg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
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

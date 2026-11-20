package main

import (
	"fmt"
	"os"

	hqeda "github.com/Huaqiu-Electronics/hqpcb-go-sdk"
	"github.com/Huaqiu-Electronics/hqpcb-go-sdk/samples/helper"
)

func main() {
	_, gui, _ := helper.ResolvePipe(os.Args)
	cli, err := hqeda.NewConnection(gui)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer cli.Disconnect()
	echo, err := cli.Ping("sample-01")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	caps, err := cli.Caps()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("pipe=%s\n", cli.Pipe())
	fmt.Printf("role=%s\n", cli.PeerRole())
	fmt.Printf("ping=%s\n", echo)
	fmt.Printf("caps interactive=%v workspace=%v design=%v ver=%d.%d\n",
		caps.Interactive, caps.Workspace, caps.Design, caps.VerMajor, caps.VerMinor)
}

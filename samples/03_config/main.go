package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/Huaqiu-Electronics/hqpcb-go-sdk"
	"github.com/Huaqiu-Electronics/hqpcb-go-sdk/samples/sampleutil"
)

func main() {
	pipe, gui, rest := sampleutil.ResolvePipe(os.Args)
	key, value := "sample.theme", "dark"
	if len(rest) > 0 {
		key = rest[0]
	}
	if len(rest) > 1 {
		value = rest[1]
	}
	cli := pcb.NewFlightIPC()
	if err := cli.Connect(pipe, gui); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer cli.Disconnect()
	out, err := cli.ConfigSet(key, value)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(strings.TrimRight(out, "\n"))
	got, err := cli.ConfigGet(key)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("%s=%s\n", key, got)
	if got != value {
		os.Exit(1)
	}
}

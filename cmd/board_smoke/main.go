// board_smoke：P6d Go IDL 绑定冒烟 — Board.open + tracks + DoGet objects.list。
//
// 用法：
//
//	board_smoke [pipe] [board.hqpcb]
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Huaqiu-Electronics/hqpcb-go-sdk"
)

func main() {
	pipe := ""
	boardPath := ""
	if len(os.Args) >= 2 {
		pipe = strings.TrimSpace(os.Args[1])
	}
	if len(os.Args) >= 3 {
		boardPath = os.Args[2]
	}
	if boardPath == "" {
		// 默认 samples 旁 complex_hierarchy。
		here, _ := os.Getwd()
		cand := filepath.Join(here, "..", "..", "..", "python", "samples", "complex_hierarchy.hqpcb")
		if st, err := os.Stat(cand); err == nil && !st.IsDir() {
			boardPath = cand
		} else {
			boardPath = filepath.Join(here, "..", "..", "..", "..", "samples", "complex_hierarchy.hqpcb")
		}
	}
	abs, err := filepath.Abs(boardPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if st, err := os.Stat(abs); err != nil || st.IsDir() {
		fmt.Fprintln(os.Stderr, "board not found:", abs)
		os.Exit(1)
	}

	cli := pcb.NewFlightIPC()
	if err := cli.Connect(pipe, false); err != nil {
		fmt.Fprintln(os.Stderr, "connect:", err)
		os.Exit(1)
	}
	defer cli.Disconnect()

	board, err := pcb.OpenBoard(pcb.BindFlightIPC(cli), abs, pcb.BindFlightDoGet(cli))
	if err != nil {
		fmt.Fprintln(os.Stderr, "open:", err)
		os.Exit(1)
	}
	tracks, err := board.Tracks(false)
	if err != nil {
		fmt.Fprintln(os.Stderr, "tracks:", err)
		os.Exit(1)
	}
	tbl, err := board.ObjectsList("tracks")
	if err != nil {
		fmt.Fprintln(os.Stderr, "objects.list:", err)
		os.Exit(1)
	}
	defer tbl.Release()
	fmt.Printf("OK pipe=%s path=%s tracks=%d doget_rows=%d ver=%d.%d\n",
		cli.Pipe(), board.Path, len(tracks), tbl.NumRows(), pcb.VerMajor, pcb.VerMinor)
}

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"

	"github.com/Huaqiu-Electronics/hqpcb-go-sdk"
	"github.com/Huaqiu-Electronics/hqpcb-go-sdk/samples/sampleutil"
)

func colNames(tbl arrow.Table) []string {
	out := make([]string, tbl.NumCols())
	for i := 0; i < int(tbl.NumCols()); i++ {
		out[i] = tbl.Schema().Field(i).Name
	}
	return out
}

func payload0(tbl arrow.Table) string {
	if tbl.NumRows() == 0 {
		return ""
	}
	names := colNames(tbl)
	idx := -1
	for i, n := range names {
		if n == "payload" {
			idx = i
			break
		}
	}
	if idx < 0 {
		return ""
	}
	col := tbl.Column(idx)
	defer col.Release()
	if col.Len() == 0 || col.Data().Len() == 0 {
		return ""
	}
	chunk := col.Data().Chunk(0)
	defer chunk.Release()
	sa, ok := chunk.(*array.String)
	if !ok {
		return chunk.ValueStr(0)
	}
	s := sa.Value(0)
	if len(s) > 120 {
		return s[:120] + "..."
	}
	return s
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
	board, err := pcb.OpenBoard(pcb.BindFlightIPC(cli), path, pcb.BindFlightDoGet(cli))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("path=%s\n", board.Path)

	tracks, err := board.ObjectsList("tracks")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer tracks.Release()
	fmt.Printf("objects.list kind=tracks rows=%d cols=%v\n", tracks.NumRows(), colNames(tracks))
	if tracks.NumRows() > 0 {
		fmt.Printf("  payload0=%s\n", payload0(tracks))
	}

	vias, err := board.ObjectsList("vias")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer vias.Release()
	fmt.Printf("objects.list kind=vias rows=%d\n", vias.NumRows())

	tbl, err := board.QueryTable("trace_segment")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer tbl.Release()
	fmt.Printf("query.table table=trace_segment rows=%d cols=%v\n", tbl.NumRows(), colNames(tbl))
	if tbl.NumRows() > 0 {
		fmt.Printf("  payload0=%s\n", payload0(tbl))
	}

	_, err = board.QueryTable("sqlite_master")
	if err == nil {
		fmt.Fprintln(os.Stderr, "error: expected table_not_allowed")
		os.Exit(2)
	}
	msg := err.Error()
	reason := "table_not_allowed"
	if !strings.Contains(msg, reason) {
		reason = strings.SplitN(msg, ".", 2)[0]
	}
	fmt.Printf("reject_ok=%T:%s\n", err, reason)
}

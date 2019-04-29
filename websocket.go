package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
	"os"
	"os/exec"
)

func streamRun(ctx context.Context, c *websocket.Conn, args ...string) bool {
	readOut, writeOut := io.Pipe()

	sail := exec.CommandContext(ctx, os.Args[0], args...)
	sail.Env = append(os.Environ(), "EDITOR=true")
	sail.Stdout = writeOut
	sail.Stderr = writeOut
	err := sail.Start()
	if err != nil {
		wsjson.Write(ctx, c, muxMsg{
			Type: "error",
			V:    fmt.Sprintf("failed to start %q: %v", sail.Args, err),
		})
		return false
	}

	go func() {
		werr := sail.Wait()
		writeOut.CloseWithError(werr)
	}()

	defer sail.Process.Kill()

	for {
		b := make([]byte, 4096)
		n, rerr := readOut.Read(b)

		if n > 0 {
			err := wsjson.Write(ctx, c, muxMsg{
				Type: "data",
				V:    b[:n],
			})
			if err != nil {
				log.Println(err)
				return false
			}
		}

		if rerr == io.EOF {
			return true
		}

		if rerr != nil {
			wsjson.Write(ctx, c, muxMsg{
				Type: "error",
				V:    fmt.Sprintf("failed to read sail output: %v", rerr),
			})
			return false
		}
	}
}

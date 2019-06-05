package sshforward

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"golang.org/x/xerrors"
)

// TODO: Make this not as bad.

type Forwarder interface {
	Forward() error
	io.Closer
}

type sshForwarder struct {
	cmd    *exec.Cmd
	cancel context.CancelFunc
}

var _ Forwarder = new(sshForwarder)

func (f *sshForwarder) Close() error {
	const closeTimeout = 5 * time.Second

	f.cancel()
	done := make(chan struct{})
	go func() {
		// Possible go routine leak if process doesn't exit within timeout.
		// Since closing usually happens at the end, this shouldn't be too big
		// of a deal.

		// TODO: handle error
		_ = f.cmd.Wait()
		done <- struct{}{}
	}()

	select {
	case <-done:
	case <-time.After(closeTimeout):
		return xerrors.Errorf("failed to close forwarder, timeout exceeded (%s)", closeTimeout)
	}
	return nil
}

func (f *sshForwarder) Forward() error {
	f.cmd.Stdout = os.Stdout
	f.cmd.Stderr = os.Stderr

	// TODO: Check that it properly forwarded instead of sleeping.
	err := f.cmd.Start()
	if err != nil {
		return xerrors.Errorf("failed to start ssh forward: %w", err)
	}

	time.Sleep(2 * time.Second)

	return nil
}

func NewLocalSocketForwarder(local string, remote string, host string) Forwarder {
	forwardFlagVal := fmt.Sprintf("%s:%s", local, remote)
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, "ssh", "-L", forwardFlagVal, "-N", host)

	return &sshForwarder{
		cmd:    cmd,
		cancel: cancel,
	}
}

func NewLocalPortForwarder(local string, remote string, host string) Forwarder {
	forwardFlagVal := fmt.Sprintf("%s:localhost:%s", local, remote)
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, "ssh", "-L", forwardFlagVal, "-N", host)

	return &sshForwarder{
		cmd:    cmd,
		cancel: cancel,
	}
}

func NewRemoteSocketForwarder(remote string, local string, host string) Forwarder {
	forwardFlagVal := fmt.Sprintf("%s:%s", remote, local)
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, "ssh", "-R", forwardFlagVal, "-N", host)

	return &sshForwarder{
		cmd:    cmd,
		cancel: cancel,
	}
}

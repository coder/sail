package main

import (
	"bytes"
	"flag"
	"os"

	"go.coder.com/cli"
	"go.coder.com/flog"
	"go.coder.com/sail/internal/dockutil"
	"go.coder.com/sail/internal/xexec"
)

type shellcmd struct {
	gf *globalFlags
}

func (c *shellcmd) Spec() cli.CommandSpec {
	return cli.CommandSpec{
		Name:  "shell",
		Desc:  "shell drops you into the default shell of a repo container.",
		Usage: "<repo>",
	}
}

func (c *shellcmd) Run(fl *flag.FlagSet) {
	proj := c.gf.project(fl)
	c.gf.ensureDockerDaemon()

	out, err := dockutil.FmtExec(proj.cntName(), "grep ^.*:.*:$(id -u): /etc/passwd | cut -d : -f 7-").CombinedOutput()
	if err != nil {
		flog.Fatal("failed to get default shell: %v\n%s", err, out)
	}

	cmd := dockutil.ExecTTY(proj.cntName(), guestHomeDir, string(bytes.TrimSpace(out)))
	xexec.Attach(cmd)
	err = cmd.Run()
	if err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

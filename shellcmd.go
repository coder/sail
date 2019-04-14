package main

import (
	"bytes"
	"flag"
	"os"

	"go.coder.com/flog"
	"go.coder.com/narwhal/internal/xexec"
)

type shellcmd struct {
}

func (c *shellcmd) spec() commandSpec {
	const desc = "shell drops you into the default shell of a repo container."
	return commandSpec{
		name:      "shell",
		shortDesc: desc,
		longDesc:  desc,
		usage:     "[repo]",
	}
}

func (c *shellcmd) handle(gf globalFlags, fl *flag.FlagSet) {
	proj := gf.project(fl)
	gf.ensureDockerDaemon()

	out, err := proj.FmtExec("grep ^.*:.*:$(id -u): /etc/passwd | cut -d : -f 7-").CombinedOutput()
	if err != nil {
		flog.Fatal("failed to get default shell: %v\n%s", err, out)
	}

	cmd := proj.ExecTTY(guestHomeDir, string(bytes.TrimSpace(out)))
	xexec.Attach(cmd)
	err = cmd.Run()
	if err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func (c *shellcmd) initFlags(fl *flag.FlagSet) {

}

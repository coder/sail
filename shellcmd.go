package main

import (
	"context"
	"flag"
	"os"
	"strings"
	"time"

	"go.coder.com/cli"
	"go.coder.com/flog"
	"go.coder.com/sail/internal/environment"
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
	repoURI := fl.Arg(0)
	if repoURI == "" {
		flog.Fatal("Argument <repo> must be provided.")
	}

	conf := c.gf.config()
	repo, err := environment.ParseRepo(conf.DefaultSchema, conf.DefaultHost, repoURI)
	if err != nil {
		flog.Fatal("failed to parse repo: %s: %v", repoURI, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	env, err := environment.FindEnvironment(ctx, repo.DockerName())
	if err != nil {
		flog.Fatal("failed to find environment: %v", err)
	}

	// Get user's login shell from /etc/passwd.
	out, err := env.Exec(ctx, "bash", "-c", "getent passwd $(whoami) | cut -d: -f7").CombinedOutput()
	if err != nil {
		flog.Fatal("failed to get user's default shell: %s: %v", out, err)
	}
	shell := strings.TrimSpace(string(out))

	cmd := env.ExecTTY(context.Background(), shell)
	xexec.Attach(cmd)
	err = cmd.Run()
	if err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

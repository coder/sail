package main

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"time"

	"go.coder.com/cli"
	"go.coder.com/flog"
	"go.coder.com/sail/internal/dockutil"
)

type rmcmd struct {
	gf *globalFlags

	repoArg  string
	all      bool
	withData bool
}

func (c *rmcmd) Spec() cli.CommandSpec {
	return cli.CommandSpec{
		Name:  "rm",
		Usage: "[flags] <repo>",
		Desc: `Remove a sail container from the system.
This command allows for removing a single container
or all of the containers on a system with the -all flag.`,
	}
}

func (c *rmcmd) RegisterFlags(fl *flag.FlagSet) {
	fl.BoolVar(&c.all, "all", false, "Remove all Sail containers.")
	fl.BoolVar(&c.withData, "with-data", false, "Remove the cloned repository's directory.")
}

func (c *rmcmd) Run(fl *flag.FlagSet) {
	c.repoArg = fl.Arg(0)

	if c.repoArg == "" && !c.all {
		fl.Usage()
		os.Exit(1)
	}

	c.gf.ensureDockerDaemon()

	names := c.getRemovalList()
	c.removeContainers(names...)
}

// getRemovalList returns a list of container names that should be removed.
func (c *rmcmd) getRemovalList() []string {
	if !c.all {
		return []string{
			toDockerName(c.repoArg),
		}
	}

	cnts, err := listContainers()
	if err != nil {
		flog.Fatal("failed to list sail containers: %v", err)
	}

	var names = make([]string, 0, len(cnts))
	for _, cnt := range cnts {
		name := trimDockerName(cnt)
		if name == "" {
			flog.Error("container %v doesn't have a name.", cnt.ID)
			continue
		}

		names = append(names, name)
	}

	return names
}

func (c *rmcmd) removeContainers(names ...string) {
	cli := dockerClient()
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	for _, name := range names {
		err := dockutil.StopRemove(ctx, cli, name)
		if err != nil {
			flog.Error("failed to remove %s: %v", name, err)
			continue
		}
		if c.withData {
			root := c.gf.config().ProjectRoot
			path := filepath.Join(root, c.repoArg)
			err = os.RemoveAll(path)
			if err != nil {
				flog.Error("Failed to remove cloned directory: %v", err)
			}
		}
		flog.Info("removed %s", name)
	}
}

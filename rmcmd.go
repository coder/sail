package main

import (
	"context"
	"flag"
	"os"
	"time"

	"go.coder.com/sail/internal/dockutil"

	"go.coder.com/flog"
)

type rmcmd struct {
	repoArg string
	all     bool
}

func (c *rmcmd) spec() commandSpec {
	return commandSpec{
		name:      "rm",
		shortDesc: "Remove a sail container from the system.",
		longDesc: `This command allows for removing a single container
	or all of the containers on a system with the -all flag.`,
		usage: "[flags] <repo>",
	}
}

func (c *rmcmd) initFlags(fl *flag.FlagSet) {
	fl.BoolVar(&c.all, "all", false, "Remove all sail containers.")
}

func (c *rmcmd) handle(gf globalFlags, fl *flag.FlagSet) {
	gf.ensureDockerDaemon()

	c.repoArg = fl.Arg(0)

	if c.repoArg == "" && !c.all {
		fl.Usage()
		os.Exit(1)
	}

	names := c.getRemovalList()

	removeContainers(names...)
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

func removeContainers(names ...string) {
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

		flog.Info("removed %s", name)
	}
}

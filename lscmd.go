package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"go.coder.com/flog"
)

type lscmd struct {
	all bool
}

func (c *lscmd) spec() commandSpec {
	return commandSpec{
		name:      "ls",
		shortDesc: "Lists all sail containers.",
		longDesc:  fmt.Sprintf(`Queries docker for all containers with the %v label.`, sailLabel),
	}
}

func (c *lscmd) initFlags(fl *flag.FlagSet) {
	fl.BoolVar(&c.all, "all", false, "Show stopped container.")
}

func (c *lscmd) handle(gf globalFlags, fl *flag.FlagSet) {
	cli := dockerClient()
	defer cli.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	filter := filters.NewArgs()
	filter.Add("label", sailLabel)

	cnts, err := cli.ContainerList(ctx, types.ContainerListOptions{
		All:     c.all,
		Filters: filter,
	})
	if err != nil {
		flog.Fatal("failed to list containers: %v", err)
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	fmt.Fprintf(tw, "name\that\turl\tstatus\n")
	for _, cnt := range cnts {
		var name string
		if len(cnt.Names) == 0 {
			// All sail containers should be named.
			flog.Error("container %v doesn't have a name.", cnt.ID)
			continue
		}
		name = strings.TrimPrefix(cnt.Names[0], "/")

		var (
			port = cnt.Labels[portLabel]
			hat  = cnt.Labels[hatLabel]
		)
		// Convert the first - into a / in order to produce a
		// sail-friendly name.
		// TODO: this is super janky.
		name = strings.Replace(name, "-", "/", 1)

		fmt.Fprintf(tw, "%v\t%v\thttp://127.0.0.1:%v\t%v\n", name, hat, port, cnt.Status)
	}
	tw.Flush()

	os.Exit(0)
}

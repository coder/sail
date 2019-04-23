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
	"golang.org/x/xerrors"
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

// projectInfo contains high-level project metadata as returned by the ls
// command.
type projectInfo struct {
	Name   string `json:"name,omitempty"`
	Hat    string `json:"hat,omitempty"`
	Url    string `json:"url,omitempty"`
	Status string `json:"status,omitempty"`
}

// listProjects grabs a list of all projects.:
func listProjects() ([]projectInfo, error) {
	cli := dockerClient()
	defer cli.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	filter := filters.NewArgs()
	filter.Add("label", sailLabel)

	cnts, err := cli.ContainerList(ctx, types.ContainerListOptions{
		All:     true,
		Filters: filter,
	})
	if err != nil {
		return nil, xerrors.Errorf("failed to list containers: %w", err)
	}

	infos := make([]projectInfo, 0, len(cnts))

	for _, cnt := range cnts {
		var info projectInfo
		if len(cnt.Names) == 0 {
			// All sail containers should be named.
			flog.Error("container %v doesn't have a name.", cnt.ID)
			continue
		}
		info.Name = strings.TrimPrefix(cnt.Names[0], "/")
		// Convert the first - into a / in order to produce a
		// sail-friendly name.
		// TODO: this is super janky.
		info.Name = strings.Replace(info.Name, "-", "/", 1)

		info.Url = "http://127.0.0.1:" + cnt.Labels[portLabel]
		info.Hat = cnt.Labels[hatLabel]

		infos = append(infos, info)
	}

	return infos, nil
}

func (c *lscmd) handle(gf globalFlags, fl *flag.FlagSet) {
	infos, err := listProjects()
	if err != nil {
		flog.Fatal("failed to list projects: %v", err)
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	fmt.Fprintf(tw, "name\that\turl\tstatus\n")
	for _, info := range infos {
		fmt.Fprintf(tw, "%v\t%v\t%v\t%v\n", info.Name, info.Hat, info.Url, info.Status)
	}
	tw.Flush()

	os.Exit(0)
}

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
	name   string
	hat    string
	url    string
	status string
}

// listProjects grabs a list of all projects.:
func listProjects() ([]projectInfo, error) {
	cnts, err := listContainers()
	if err != nil {
		return nil, xerrors.Errorf("failed to list containers: %w", err)
	}

	infos := make([]projectInfo, 0, len(cnts))

	for _, cnt := range cnts {
		var info projectInfo

		dockerName := trimDockerName(cnt)
		if dockerName == "" {
			flog.Error("container %v doesn't have a name.", cnt.ID)
			continue
		}
		info.name = toSailName(dockerName)

		url, err := proxyURL(dockerName)
		if err != nil {
			return nil, xerrors.Errorf("failed to find container %s port: %w", info.name, err)
		}
		info.url = url
		info.hat = cnt.Labels[hatLabel]

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
		fmt.Fprintf(tw, "%v\t%v\t%v\t%v\n", info.name, info.hat, info.url, info.status)
	}
	tw.Flush()

	os.Exit(0)
}

// listContainers lists the sail containers on the host that
// are filterable by the sail label: com.coder.sail
func listContainers() ([]types.Container, error) {
	cli := dockerClient()
	defer cli.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	filter := filters.NewArgs()
	filter.Add("label", sailLabel)

	return cli.ContainerList(ctx, types.ContainerListOptions{
		All:     true,
		Filters: filter,
	})
}

// trimDockerName trims the `/` prefix from the docker container name.
// If the container isn't named, this will return the empty string.
func trimDockerName(cnt types.Container) string {
	if len(cnt.Names) == 0 {
		// All sail containers should be named.
		return ""
	}
	return strings.TrimPrefix(cnt.Names[0], "/")
}

// toSailName converts the first _ into a / in order to produce a
// sail-friendly name.
//
// TODO: this is super janky.
func toSailName(dockerName string) string {
	return strings.Replace(dockerName, "_", "/", 1)
}

// toDockerName converts the first / into a _ in order to produce
// a docker-friendly name.
//
// TODO: this is super janky.
func toDockerName(sailName string) string {
	return strings.Replace(sailName, "/", "_", 1)
}

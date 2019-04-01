package main

import (
	"flag"
	"go.coder.com/flog"
	"go.coder.com/narwhal/internal/randstr"
	"golang.org/x/xerrors"
	"os"
	"strings"
	"time"
)

// run is the entry-point to the narwhal command.
func run(flg flags, c Config) {
	//flog.Info("config: %+v", c)

	hostname := flag.Arg(0)
	repo := flag.Arg(1)

	if repo == "" {
		repo = flag.Arg(0)
		hostname = c.DefaultHost
	}

	if repo == "" {
		flog.Fatal("Argument <repo> must be provided.")
	}

	flg.debug("host: %v, repo: %v", hostname, repo)

	host := Dial(hostname)

	out, err := host.Command("docker", "info").CombinedOutput()
	if err != nil {
		flog.Fatal("failed to run `docker info`: %v\n%s", err, out)
	}
	flg.debug("verified Docker is running on %v", hostname)

	r, err := ParseRepo(repo)
	if err != nil {
		flog.Fatal("failed to parse repo %q: %v", repo, err)
	}

	basename := r.DockerName()

	var (
		name string
	)

	if !flg.newContainer {
		// Try to re-use an existing container.
		names, err := ListContainers(host, basename)
		if err != nil {
			flog.Fatal("failed to list containers: %v", err)
		}
		if len(names) > 0 {
			// First container is the most recently created one.
			name = names[0]
			flog.Info("reusing container %v", name)
		}
	}
	if name == "" {
		name = basename + "-" + strings.ToLower(randstr.Make(6))
		err = RunContainer(os.Stderr, host, c.DefaultImage, name)
		if err != nil {
			flog.Fatal("failed to run container: %v", err)
		}
		flg.debug("Started container %v", name)
	}

	cnt := container{Name: name, Host: host}

	start := time.Now()
	err = cnt.DownloadCodeServer()
	if err != nil {
		flog.Fatal("failed to download code-server: %v", err)
	}
	flg.debug("downloaded code-server in %v", time.Since(start))

	err = cnt.StartCodeServer("8080")
	if err != nil {
		if xerrors.Is(err, errCodeServerFailed) {
			log, err := cnt.ReadCodeServerLog()
			if err != nil {
				flog.Fatal("code-server failed to start, and couldn't read logs: %v", err)
			}
			flog.Fatal("code-server failed to start.\n%s", log)
		}
		flog.Fatal("failed to start code-server: %v", err)
	}
}

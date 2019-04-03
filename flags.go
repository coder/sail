package main

import (
	"flag"
	"github.com/fatih/color"
	"github.com/skratchdot/open-golang/open"
	"go.coder.com/flog"
	"golang.org/x/xerrors"
	"net"
	"os"
)

type flags struct {
	verbose    bool
	configPath string
	prune      bool
	image      string
	hat        string
	shell      bool
	exec       bool
}

func initFlags() *flags {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	flg := new(flags)
	flag.BoolVar(&flg.verbose, "v", false, "Enable debug logging.")
	flag.StringVar(&flg.configPath, "config", homeDir+"/.config/narwhal/narwhal.toml", "Path to config.")
	flag.BoolVar(&flg.prune, "prune", false, "Remove all narwhal containers from host.")
	flag.StringVar(&flg.image, "image", "", "Docker image to use.")
	flag.StringVar(&flg.hat, "hat", "", "Path to hat.")
	flag.BoolVar(&flg.exec, "exec", false, "Execute a command in the container.")
	flag.BoolVar(&flg.shell, "shell", false, "Drop into container's shell.")
	flag.BoolVar(&flg.shell, "s", false, "Shorthand for --shell.")
	return flg
}

func (flg *flags) debug(msg string, args ...interface{}) {
	if !flg.verbose {
		return
	}
	flog.Log(
		flog.Level(color.New(color.FgHiMagenta).Sprint("DEBUG")),
		msg, args...,
	)
}

// run is the entry-point to the narwhal command.
func (flg *flags) run(c config) {
	if flg.prune {
		os.Exit(handlePrune(flg))
	}

	repo := flag.Arg(0)

	if repo == "" {
		flog.Fatal("Argument <repo> must be provided.")
	}

	flg.ensureDockerDaemon()

	r, err := ParseRepo(repo)
	if err != nil {
		flog.Fatal("failed to parse repo %q: %v", repo, err)
	}

	name := r.DockerName()

	// Try to re-use an existing container.
	exists, err := containerExists(name)
	if err != nil {
		flog.Fatal("failed to inspect existing container %v: %v", name, err)
	}
	if exists {
		flg.debug("Reusing container %v", name)
		cmd := fmtExec("docker start %v", name)
		out, err := cmd.CombinedOutput()
		if err != nil {
			flog.Fatal("failed to start container: %v\n%s", err, out)
		}
	} else {
		codeServerBinPath, err := loadCodeServer(context.Background())
		if err != nil {
			flog.Fatal("failed to load code-server: %v", err)
		}
		c.Shares[codeServerBinPath] = "/usr/bin/code-server"

		projectDir := projectDir(cleanPath(c.ProjectRoot), r)
		c.Shares[ensureProject(projectDir, r)] = c.containerProjectPath(r)

		var image string
		if flg.image != "" {
			image = flg.image
		} else {
			var customImageExists bool
			image, customImageExists = ensureProjectImage(projectDir, r)
			if !customImageExists {
				flog.Info("using default image %v", c.DefaultImage)
				image = c.DefaultImage
			} else {
				flog.Info("using repo image %v", image)
			}
		}

		image = applyHat(flg, c, image)

		populateImageShares(image, c.Shares)

		err = runContainer(os.Stderr, containerConfig{
			image:    image,
			name:     name,
			hostname: r.BaseName(),
			shares:   c.Shares,
		})
		if err != nil {
			flog.Fatal("failed to run container: %v", err)
		}
		flg.debug("Started container %v", name)
	}

	cnt := container{Name: name}

	if flg.shell {
		os.Exit(handleShell(r, cnt))
	}
	if flg.exec {
		os.Exit(handleExec(r, cnt))
	}

	out, err := cnt.FmtExec("mkdir -p %v", c.containerProjectPath(r)).CombinedOutput()
	if err != nil {
		flog.Fatal("failed to create %v: %v\n%s", c.containerProjectPath(r), err, out)
	}

	port, err := findAvailablePort(8000, 9000)
	if err != nil {
		flog.Fatal("failed to find available port: %v", err)
	}

	err = cnt.StartCodeServer(c.containerProjectPath(r), port)
	if err != nil {
		switch {
		case xerrors.Is(err, errCodeServerFailed):
			log, err := cnt.ReadCodeServerLog()
			if err != nil {
				flog.Fatal("code-server failed to start, and couldn't read logs: %v", err)
			}
			flog.Fatal("code-server failed to start.\n%s", log)
		case xerrors.Is(err, errCodeServerRunning):
			port, err = cnt.CodeServerPort()
			if err != nil {
				flog.Fatal("code-server is running, but can't detect port: %v", err)
			}
			flg.debug("found code-server running on port %v", port)
		default:
			flog.Fatal("failed to start code-server: %v", err)
		}
	} else {
		flg.debug("code-server started successfully")
	}

	u := "http://" + net.JoinHostPort("127.0.0.1", port)
	flog.Info("opening browser serving %v", u)
	_ = open.Run(u)
}

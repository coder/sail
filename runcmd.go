package main

import (
	"flag"
	"github.com/docker/docker/api/types"
	"go.coder.com/flog"
	"golang.org/x/xerrors"
	"os"
)

type runcmd struct {
	repoArg string

	image string
	hat   string
}

func (c *runcmd) spec() commandSpec {
	return commandSpec{
		name:      "run",
		shortDesc: "Runs a project container.",
		longDesc:  ``,
		usage:     "[project]",
	}
}

func (c *runcmd) initFlags(fl *flag.FlagSet) {
	c.repoArg = fl.Arg(0)
	fl.StringVar(&c.image, "image", "", "Custom docker baseImage to use.")
	fl.StringVar(&c.hat, "hat", "", "Custom hat to use.")
}

func (c *runcmd) handle(gf globalFlags, fl *flag.FlagSet) {
	var (
		err error
	)
	gf.ensureDockerDaemon()

	proj := gf.project(fl)

	// Abort if container already exists.
	exists := proj.cntExists()
	if exists {
		flog.Fatal(
			"Container %v already exists. Use `nw open %v` to open it.",
			proj.cntName(), proj.cntName(),
		)
	}

	proj.ensureDir()

	var shares []types.MountPoint
	shares = append(shares, types.MountPoint{
		Type:        "bind",
		Source:      proj.dir(),
		Destination: proj.containerDir(),
	})

	var image string
	if c.image != "" {
		image = c.image
	} else {
		var customImageExists bool
		image, customImageExists, err = proj.buildImage()
		if err != nil {
			flog.Fatal("failed to build image: %v", err)
		}
		if !customImageExists {
			flog.Info("using default baseImage %v", gf.config().DefaultImage)
			image = gf.config().DefaultImage
		} else {
			flog.Info("using repo image %v", image)
		}
	}

	// Apply hat if configured.
	var hatPath string
	switch {
	case c.hat != "":
		hatPath = c.hat
	case gf.config().DefaultHat != "":
		hatPath = gf.config().DefaultHat
	}

	for k, v := range gf.config().Shares {
		shares = append(shares, types.MountPoint{
			Type:        "bind",
			Source:      cleanPath(k),
			Destination: v,
		})
	}

	b := &builder{
		baseImage: image,
		hatPath:   hatPath,
		name:      proj.cntName(),
		hostname:  proj.repo.BaseName(),
		shares:    shares,
	}
	err = b.runContainer()
	if err != nil {
		flog.Fatal("failed to run container: %v", err)
	}

	gf.debug("Started container %v", proj.cntName())

	port, err := findAvailablePort()
	if err != nil {
		flog.Fatal("failed to find available port: %v", err)
	}

	err = proj.StartCodeServer(proj.containerDir(), port)
	if err != nil {
		switch {
		case xerrors.Is(err, errCodeServerFailed):
			log, err := proj.ReadCodeServerLog()
			if err != nil {
				flog.Fatal("code-server failed to start, and couldn't read logs: %v", err)
			}
			flog.Fatal("code-server failed to start.\n%s", log)
		case xerrors.Is(err, errCodeServerRunning):
			port, err = proj.CodeServerPort()
			if err != nil {
				flog.Fatal("code-server is running, but can't detect port: %v", err)
			}
			gf.debug("found code-server running on port %v", port)
		default:
			flog.Fatal("failed to start code-server: %v", err)
		}
	} else {
		gf.debug("code-server started successfully")
	}

	err = proj.open()
	if err != nil {
		flog.Fatal("failed to open project: %v", err)
	}
	os.Exit(0)
}

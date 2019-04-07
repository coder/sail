package main

import (
	"context"
	"flag"
	"github.com/docker/docker/api/types"
	"go.coder.com/flog"
	"go.coder.com/narwhal/internal/dockutil"
	"golang.org/x/xerrors"
	"os"
)

type runcmd struct {
	repoArg string

	image string
	hat   string
	keep  bool
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
	fl.BoolVar(&c.keep, "keep", false, "Keep container when it fails to build.")
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
			proj.cntName(), proj.name(),
		)
	}

	proj.ensureDir()

	var shares []types.MountPoint
	shares = append(shares, types.MountPoint{
		Type:        "bind",
		Source:      proj.localDir(),
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

	port, err := findAvailablePort()
	if err != nil {
		flog.Fatal("failed to find available port: %v", err)
	}

	b := &builder{
		baseImage:  image,
		hatPath:    hatPath,
		name:       proj.cntName(),
		hostname:   proj.repo.BaseName(),
		shares:     shares,
		port:       port,
		projectDir: proj.containerDir(),
	}

	err = c.buildOpen(gf, proj, b)
	if err != nil {
		if !c.keep {
			// We remove the container if it fails to start as that means the developer
			// can iterate w/o having to do the obnoxious `docker rm` step.
			gf.debug("removing %v", proj.cntName())
			err = dockutil.StopRemove(context.Background(), dockerClient(), proj.cntName())
			if err != nil {
				flog.Error("failed to remove %v", proj.cntName())
			}
		}
		os.Exit(1)
	}
	os.Exit(0)
}

func (c *runcmd) buildOpen(gf globalFlags, proj *project, b *builder) error {
	var err error
	err = b.runContainer()
	if err != nil {
		return xerrors.Errorf("failed to run container: %w", err)
	}

	gf.debug("started container")

	err = proj.waitOnline()
	if err != nil {
		flog.Error("failed to wait for project to be online: %v", err)

		logs, logErr := proj.readCodeServerLog()
		if logErr != nil {
			return xerrors.Errorf("%v\nfailed to read log: %w", err, logErr)
		}
		os.Stderr.Write(logs)

		return err
	}

	gf.debug("code-server online")

	err = proj.open()
	if err != nil {
		return xerrors.Errorf("failed to open project: %w", err)
	}
	return nil
}

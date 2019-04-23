package main

import (
	"context"
	"flag"
	"os"
	"os/user"

	"go.coder.com/flog"
	"go.coder.com/sail/internal/dockutil"
	"golang.org/x/xerrors"
)

type runcmd struct {
	repoArg string

	image   string
	hat     string
	keep    bool
	testCmd string
}

func (c *runcmd) spec() commandSpec {
	return commandSpec{
		name:      "run",
		shortDesc: "Runs a project container.",
		longDesc: `This command is used for opening and running a project.
	If a project is not yet created or running with the name,
	one will be created and a new editor will be opened.
	If a project is already up and running, this won't
	start a new container, but instead will reuse the
	already running container and open a new editor.`,
		usage: "[flags] <repo>",
	}
}

func (c *runcmd) initFlags(fl *flag.FlagSet) {
	c.repoArg = fl.Arg(0)

	fl.StringVar(&c.image, "image", "", "Custom docker image to use.")
	fl.StringVar(&c.hat, "hat", "", "Custom hat to use.")
	fl.BoolVar(&c.keep, "keep", false, "Keep container when it fails to build.")
	fl.StringVar(&c.testCmd, "test-cmd", "", "A command to use in-place of starting code-server for testing purposes.")
}

const guestHomeDir = "/home/user"

func (c *runcmd) handle(gf globalFlags, fl *flag.FlagSet) {
	gf.ensureDockerDaemon()
	proj := gf.project(fl)

	exists, err := proj.cntExists()
	if err != nil {
		flog.Fatal("%v", err)
	}
	if !exists {
		c.build(gf, fl)
	}

	err = proj.open()
	if err != nil {
		flog.Error("failed to open project: %v", err)
		err = proj.delete()
		if err != nil {
			flog.Error("failed to delete project container: %v", err)
		}
		os.Exit(1)
	}
}

func (c *runcmd) build(gf globalFlags, fl *flag.FlagSet) {
	proj := gf.project(fl)
	err := proj.ensureDir()
	if err != nil {
		flog.Fatal("%v", err)
	}

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
			flog.Info("using default image %v", gf.config().DefaultImage)
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

	hostHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	gf.debug("host home dir: %v", hostHomeDir)

	u, err := user.Current()
	if err != nil {
		flog.Fatal("failed to get current user: %v", err)
	}

	b := &hatBuilder{
		baseImage: image,
		hatPath:   hatPath,
	}

	r := &runner{
		projectName:     proj.repo.BaseName(),
		projectLocalDir: proj.localDir(),
		cntName:         proj.cntName(),
		hostname:        proj.repo.BaseName(),
		// Use `0` as the port so that the host assigns an available one.
		port:     "0",
		hostUser: u.Uid,
		testCmd:  c.testCmd,
	}

	err = c.run(gf, proj, b, r)
	if err != nil {
		flog.Error("build run failed: %v", err)
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
}

func (c *runcmd) run(gf globalFlags, proj *project, b *hatBuilder, r *runner) error {
	var err error
	image := b.baseImage
	if b.hatPath != "" {
		image, err = b.applyHat()
		if err != nil {
			return err
		}
	}

	err = r.runContainer(image)
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

	return nil
}

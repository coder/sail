package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"time"

	"golang.org/x/xerrors"

	"go.coder.com/cli"
	"go.coder.com/flog"
	"go.coder.com/sail/internal/dockutil"
)

type runcmd struct {
	gf *globalFlags

	image   string
	hat     string
	keep    bool
	testCmd string

	schemaPrefs
}

type schemaPrefs struct {
	ssh   bool
	http  bool
	https bool
}

func (c *runcmd) Spec() cli.CommandSpec {
	return cli.CommandSpec{
		Name:  "run",
		Usage: "[flags] <repo>",
		Desc: `Runs a project container.
If a project is not yet created or running with the name,
one will be created and a new editor will be opened.
If a project is already up and running, this won't
start a new container, but instead will reuse the
already running container and open a new editor.

If a schema and host are not provided, sail will use github over SSH.
There are multiple ways to modify this behavior.

1. Specify a host. See examples section
2. Specify a schema and host. See examples section
3. Edit the config to provide your preferred defaults.

Examples:
	Use default host and schema (github.com over SSH, editable in config)
	- sail run cdr/code-server

	Force SSH on a Github repo (user git is assumed by default)
	- sail run ssh://github.com/cdr/sshcode
	- sail run --ssh github.com/cdr/sshcode

	Specify a custom SSH user
	- sail run ssh://colin@git.colin.com/super/secret-repo
	- sail run --ssh colin@git.colin.com/super/secret-repo

	Force HTTPS on a Gitlab repo
	- sail run https://gitlab.com/inkscape/inkscape
	- sail run --https gitlab.com/inkscape/inkscape
	
Note:
If you use ssh://, http://, or https://, you must specify a host. 

This won't work:
	- sail run ssh://cdr/code-server

Instead, use flags to avoid providing a host.

This will work:
	- sail run --ssh cdr/code-server`,
	}
}

func (c *runcmd) RegisterFlags(fl *flag.FlagSet) {
	fl.StringVar(&c.image, "image", "", "Custom docker image to use.")
	fl.StringVar(&c.hat, "hat", "", "Custom hat to use.")
	fl.BoolVar(&c.keep, "keep", false, "Keep container when it fails to build.")
	fl.StringVar(&c.testCmd, "test-cmd", "", "A command to use in-place of starting code-server for testing purposes.")

	fl.BoolVar(&c.ssh, "ssh", false, "Clone repo over SSH")
	fl.BoolVar(&c.http, "http", false, "Clone repo over HTTP")
	fl.BoolVar(&c.https, "https", false, "Clone repo over HTTPS")
}

const guestHomeDir = "/home/user"

func (c *runcmd) Run(fl *flag.FlagSet) {
	c.gf.ensureDockerDaemon()

	proj := c.gf.project(c.schemaPrefs, fl)

	// Abort if container already exists.
	exists, err := proj.cntExists()
	if err != nil {
		flog.Fatal("%v", err)
	}
	if exists {
		c.gf.debug("opening existing project")

		u, err := proj.proxyURL()
		if err != nil {
			flog.Fatal("%v", err)
		}

		resp, err := http.Get(u + "/sail/api/v1/heartbeat")
		if err == nil {
			resp.Body.Close()

			err = proj.open()
			if err != nil {
				flog.Error("failed to open project: %v", err)
				err = proj.delete()
				if err != nil {
					flog.Error("failed to delete project container: %v", err)
				}
				os.Exit(1)
			}
			os.Exit(0)
		}

		// Proxy is not up, meaning the container shut down at some point, or the proxy
		// was killed. We're going to restart the proxy and update the container label.

		cli := dockerClient()
		defer cli.Close()

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		err = dockutil.StopRemove(ctx, cli, proj.cntName())
		if err != nil {
			flog.Fatal("failed to remove container without running proxy: %v", err)
		}

		// The container will be rebuilt properly.
	}

	err = proj.ensureDir()
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
			image = proj.defaultRepoImage()
			flog.Info("using default image %v", image)

			err = ensureImage(image)
			if err != nil {
				flog.Fatal("failed to ensure image %v: %v", image, err)
			}
		} else {
			flog.Info("using repo image %v", image)
		}
	}

	// Apply hat if configured.
	var hatPath string
	switch {
	case c.hat != "":
		hatPath = c.hat
	case c.gf.config().DefaultHat != "":
		hatPath = c.gf.config().DefaultHat
	}

	hostHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	c.gf.debug("host home dir: %v", hostHomeDir)

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
		port:    "0",
		testCmd: c.testCmd,
	}

	err = c.buildOpen(c.gf, proj, b, r)
	if err != nil {
		flog.Error("build run failed: %v", err)
		if !c.keep {
			// We remove the container if it fails to start as that means the developer
			// can iterate w/o having to do the obnoxious `docker rm` step.
			c.gf.debug("removing %v", proj.cntName())
			err = dockutil.StopRemove(context.Background(), dockerClient(), proj.cntName())
			if err != nil {
				flog.Error("failed to remove %v", proj.cntName())
			}
		}
		os.Exit(1)
	}
	os.Exit(0)
}

func (c *runcmd) buildOpen(gf *globalFlags, proj *project, b *hatBuilder, r *runner) error {
	var err error
	image := b.baseImage
	if b.hatPath != "" {
		image, err = b.applyHat()
	}

	// TODO proxy if container already exists.
	err = r.forkProxy()
	if err != nil {
		return xerrors.Errorf("failed to start proxy: %w", err)
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

	err = proj.open()
	if err != nil {
		return xerrors.Errorf("failed to open project: %w", err)
	}
	return nil
}

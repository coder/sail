package main

import (
	"context"
	"flag"
	"github.com/docker/docker/api/types"
	"go.coder.com/flog"
	"go.coder.com/narwhal/internal/dockutil"
	"go.coder.com/narwhal/internal/editor"
	"go.coder.com/narwhal/internal/randstr"
	"go.coder.com/narwhal/internal/xexec"
	"golang.org/x/xerrors"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type editcmd struct {
}

func (c *editcmd) spec() commandSpec {
	return commandSpec{
		name:      "edit",
		shortDesc: "edit your environment in real-time.",
		longDesc: `This command drops you into
your default editor, with the repo's Dockerfile open. When the editor is closed, the environment
is re-created to spec.'`,
		usage: "[repo]",
	}
}

func (c *editcmd) handle(gf globalFlags, fl *flag.FlagSet) {
	proj := gf.project(fl)

	gf.ensureDockerDaemon()

	err := os.MkdirAll(filepath.Dir(proj.dockerfilePath()), 0755)
	if err != nil {
		flog.Fatal("failed to create intermediate dirs: %v", err)
	}

	// Create file if it doesn't already exist.
	fi, err := os.OpenFile(proj.dockerfilePath(), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0640)
	if err != nil && !os.IsExist(err) {
		flog.Fatal("failed to open %v: %v", proj.dockerfilePath())
	} else if err == nil {
		defer fi.Close()
		// Provide a sensible default Dockerfile if the image hasn't been customized.
		_, err = fi.WriteString("FROM codercom/ubuntu-dev\n")
		if err != nil {
			flog.Fatal("failed to write default Dockerfile: %v", err)
		}
		err = fi.Close()
		if err != nil {
			flog.Fatal("failed to write default Dockerfile")
		}
	}

	editor, err := editor.Env()
	if err != nil {
		flog.Fatal("failed to get editor: %v", err)
	}
	// TODO: in an ideal world we could re-build the environment on each save instead of when the environment
	// quits. The problem is user feedback. For real-time edits, we must either send build feedback to the
	// calling terminal or start the editor in a completely different terminal. In the former option,
	// build feedback corrupts a terminal editor's layout. In the latter option, compatibility becomes
	// difficult, and narwhal will have a hard time being ran on server.

	cmd := exec.Command(editor, proj.dockerfilePath())
	xexec.Attach(cmd)

	err = cmd.Run()
	if err != nil {
		os.Exit(1)
	}
	err = c.recreate(proj)
	if err != nil {
	}
	os.Exit(0)
}

func (c *editcmd) recreate(proj *project) error {
	cli := dockerClient()
	defer cli.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Get port.
	var (
		port string
		err  error
	)
	if !proj.running() {
		port, err = findAvailablePort()
		if err != nil {
			return err
		}
	} else {
		port, err = proj.CodeServerPort()
		if err != nil {
			return err
		}
	}

	// Get the existing container's state so re-create is seamless.
	b := builderFromContainer(proj.cntName())

	// We move the existing container to a temporary name so that the old environment can be recovered if
	// the new environment is broken.
	tmpCntName := proj.cntName() + "-tmp-" + randstr.Make(5)

	// Rename existing container with intention of deleting it if everything goes smoothly.
	err = cli.ContainerRename(ctx, proj.cntName(), tmpCntName)
	if err != nil {
		flog.Fatal("failed to rename %v to %v: %v", proj.cntName(), tmpCntName, err)
	}

	defer func() {
		// Remove the old temporary container.
		_ = dockutil.StopRemove(ctx, cli, tmpCntName)
	}()

	// Build new image and container, and rollback if it doesn't go well.
	err = func() error {
		image, ok, err := proj.buildImage()
		if err != nil {
			return err
		}
		// If we were previously using the default image, we must explicitely override
		// to use the new base.
		if ok {
			b.baseImage = image
		}

		// Stop OG container after image is built so the period of downtime is minimized.
		err = cli.ContainerStop(ctx, tmpCntName, dockutil.DurationPtr(time.Second))
		if err != nil {
			return err
		}

		err = b.runContainer()
		if err != nil {
			return err
		}

		return nil
	}()
	if err != nil {
		flog.Error("failed to build and run new container: %v", err)
		flog.Info("rolling back...")

		// If the new, broken container exists in any shape or form, delete it.
		_ = dockutil.StopRemove(ctx, cli, proj.cntName())

		err = cli.ContainerRename(ctx, tmpCntName, proj.cntName())
		if err != nil {
			flog.Fatal("failed to rename %v to %v: %v", tmpCntName, proj.cntName(), err)
		}
	}

	// (Idempotent)
	err = cli.ContainerStart(ctx, proj.cntName(), types.ContainerStartOptions{})
	if err != nil {
		return xerrors.Errorf("failed to start container: %w", err)
	}

	err = proj.StartCodeServer(proj.containerDir(), port)
	if err != nil {
		return xerrors.Errorf("failed to start code-server: %w", err)
	}
	flog.Info("built new container")
	return nil
}

func (c *editcmd) initFlags(fl *flag.FlagSet) {

}

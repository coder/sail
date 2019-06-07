package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"time"

	"golang.org/x/xerrors"

	"go.coder.com/cli"
	"go.coder.com/flog"
	"go.coder.com/sail/internal/environment"
)

type runcmd struct {
	gf *globalFlags

	image   string
	hat     string
	keep    bool
	testCmd string

	schemaPrefs

	rebuild bool
	noOpen  bool
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
	fl.BoolVar(&c.rebuild, "rebuild", false, "Delete existing container")
	fl.BoolVar(&c.noOpen, "no-open", false, "Don't open an editor session")
}

const guestHomeDir = "/home/user"

func (c *runcmd) Run(fl *flag.FlagSet) {
	if c.gf.remoteHost != "" {
		closeSockFn, err := environment.EnsureRemoteContainerSock(c.gf.remoteHost)
		if err != nil {
			flog.Fatal("failed to ensure remote container socket: %v", err)
		}
		defer closeSockFn()
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	repoURI := fl.Arg(0)
	if repoURI == "" {
		flog.Fatal("Argument <repo> must be provided.")
	}

	conf := c.gf.config()
	schema := defaultSchema(conf, c.schemaPrefs)
	repo, err := environment.ParseRepo(schema, conf.DefaultHost, repoURI)
	if err != nil {
		flog.Fatal("failed to parse repo %s: %v", repoURI, err)
	}

	name := repo.DockerName()
	_, err = environment.FindEnvironment(ctx, name)
	if xerrors.Is(err, environment.ErrMissingContainer) {
		buildCfg := environment.NewDefaultBuildConfig(name)
		_, err = environment.Bootstrap(ctx, buildCfg, &repo)
		if err != nil {
			flog.Fatal("failed to bootstrap environment: %v", err)
		}
	} else if err != nil {
		flog.Fatal("failed to find environment: %v", err)
	}

	flog.Info("running...")

	// Keep the docker socket forwarded.
	// TODO: How should this be handled?
	if c.gf.remoteHost != "" {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, os.Interrupt, os.Kill)
		<-sigs
		flog.Info("closing...")
	}
}

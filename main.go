package main

import (
	"flag"
	"path/filepath"

	"go.coder.com/cli"
)

// A dedication to Nhooyr Software.
var _ interface {
	cli.Command
	cli.FlaggedCommand
	cli.ParentCommand
} = new(rootCmd)

type rootCmd struct {
	globalFlags
}

func (r *rootCmd) Spec() cli.CommandSpec {
	return cli.CommandSpec{
		Name:  "sail",
		Usage: "[GLOBAL FLAGS] COMMAND [COMMAND FLAGS] [ARGS....]",
		Desc: `A utility for managing Docker-based code-server environments.
More info: https://github.com/codercom/sail

[project] can be of form <org>/<repo> for GitHub repos, or the full git clone address.`,
	}
}

func (r *rootCmd) Run(fl *flag.FlagSet) {
	// The root command doesn't do anything.
	fl.Usage()
}

func (r *rootCmd) RegisterFlags(fl *flag.FlagSet) {
	fl.BoolVar(&r.verbose, "v", false, "Enable debug logging.")
	fl.StringVar(&r.configPath, "config",
		filepath.Join(metaRoot(), "sail.toml"),
		"Path to config.",
	)
}

func (r rootCmd) Subcommands() []cli.Command {
	return []cli.Command{
		&runcmd{gf: &r.globalFlags},
		&shellcmd{gf: &r.globalFlags},
		&editcmd{gf: &r.globalFlags},
		&lscmd{},
		&rmcmd{gf: &r.globalFlags},
		&proxycmd{},
	}
}

func main() {
	root := &rootCmd{}
	cli.RunRoot(root)
}

package main

import (
	"flag"
	"github.com/posener/complete"
	"os"
	"path/filepath"
	"strings"

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

	// We don't use these directly, just added for visability on fl.Usage().
	fl.Bool("install-autocomplete", false, "Install autocomplete")
	fl.Bool("uninstall-autocomplete", false, "Uninstall autocomplete")
}

func (r rootCmd) Subcommands() []cli.Command {
	return []cli.Command{
		&runcmd{gf: &r.globalFlags},
		&shellcmd{gf: &r.globalFlags},
		&editcmd{gf: &r.globalFlags},
		&lscmd{},
		&rmcmd{gf: &r.globalFlags},
		&proxycmd{},
		&extCmd{},
		&chromeExtInstall{},
	}
}

func main() {
	root := &rootCmd{}

	if handleAutocomplete(root) {
		return
	}

	if len(os.Args) >= 2 && strings.HasPrefix("chrome-extension://", os.Args[1]) ||
		len(os.Args) >= 3 && strings.HasPrefix("chrome-extension://", os.Args[2]) {
		runNativeMsgHost()
		return
	}

	cli.RunRoot(root)
}

func handleAutocomplete(root interface {
	cli.Command
	cli.ParentCommand
	cli.FlaggedCommand
}) bool {
	cmds := []cli.Command{root}
	cmds = append(cmds, root.Subcommands()...)

	cmp := complete.New("sail", genAutocomplete(cmds))
	cmp.InstallName = "install-autocomplete"
	cmp.UninstallName = "uninstall-autocomplete"
	return cmp.Run()
}

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

	installAutocomplete   bool
	uninstallAutocomplete bool
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
	if r.handleAutocomplete() {
		return
	}

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
	fl.BoolVar(&r.installAutocomplete, "install-autocomplete", false, "Install autocomplete")
	fl.BoolVar(&r.uninstallAutocomplete, "uninstall-autocomplete", false, "Uninstall autocomplete")
}

func (r rootCmd) Subcommands() []cli.Command {
	return []cli.Command{
		&runcmd{gf: &r.globalFlags},
		&shellcmd{gf: &r.globalFlags},
		&editcmd{gf: &r.globalFlags},
		&lscmd{},
		&rmcmd{gf: &r.globalFlags},
		&proxycmd{},
		&chromeExtInstall{},
	}
}

func main() {
	root := &rootCmd{}

	if len(os.Args) >= 2 && strings.HasPrefix("chrome-extension://", os.Args[1]) ||
		len(os.Args) >= 3 && strings.HasPrefix("chrome-extension://", os.Args[2]) {
		runNativeMsgHost()
		return
	}

	cli.RunRoot(root)
}

func (r *rootCmd) handleAutocomplete() bool {
	cmds := []cli.Command{r}
	cmds = append(cmds, cli.ParentCommand(r).Subcommands()...)

	cmp := complete.New("sail", genAutocomplete(cmds))
	cmp.InstallName = "install-autocomplete"
	cmp.UninstallName = "uninstall-autocomplete"

	// only call run if we know we want to install/uninstall autocomplete
	if r.installAutocomplete || r.uninstallAutocomplete {
		return cmp.Run()
	}

	// otherwise just process autocomplete
	return cmp.Complete()
}

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.coder.com/flog"
)

func flagHelp(fs *flag.FlagSet) string {
	var bd strings.Builder
	fmt.Fprintf(&bd, "Flags:\n")
	var count int
	fs.VisitAll(func(f *flag.Flag) {
		count++
		if f.DefValue == "" {
			fmt.Fprintf(&bd, "\t-%v\t%v\n", f.Name, f.Usage)
		} else {
			fmt.Fprintf(&bd, "\t-%v\t%v\t(%v)\n", f.Name, f.Usage, f.DefValue)
		}
	})
	if count == 0 {
		return "\n"
	}
	return bd.String()
}

func main() {
	var gf globalFlags

	gfs := flag.NewFlagSet("global", flag.ExitOnError)

	gfs.BoolVar(&gf.verbose, "v", false, "Enable debug logging.")
	gfs.StringVar(&gf.configPath, "config",
		filepath.Join(metaRoot(), "sail.toml"),
		"Path to config.",
	)

	cmds := []command{
		new(runcmd),
		new(shellcmd),
		new(editcmd),
		new(lscmd),
		new(rmcmd),
		new(extCmd),
	}

	gfs.Usage = func() {
		var commandHelp strings.Builder
		for _, cmd := range cmds {
			fmt.Fprintf(&commandHelp, "\t%v\t%v\n", cmd.spec().name, cmd.spec().shortDesc)
		}

		fmt.Printf(`Usage: %v [GLOBAL FLAGS] COMMAND [COMMAND FLAGS] [ARGS....]

A utility for managing Docker-based code-server environments.
More info: https://github.com/codercom/sail

[project] can be of form <org>/<repo> for GitHub repos, or the full git clone address.

Global %v
Commands:
%v
`, os.Args[0], flagHelp(gfs), commandHelp.String())
	}
	_ = gfs.Parse(os.Args[1:])

	wantCmd := gfs.Arg(0)
	if wantCmd == "" {
		gfs.Usage()
		os.Exit(1)
	}

	// help indicates if we're trying to access a command's help.
	var help bool
	if wantCmd == "help" {
		help = true
		wantCmd = gfs.Arg(1)
		if wantCmd == "" {
			gfs.Usage()
			os.Exit(0)
		}
	}

	for _, cmd := range cmds {
		if wantCmd != cmd.spec().name {
			continue
		}
		fs := flag.NewFlagSet(cmd.spec().name, flag.ExitOnError)
		cmd.initFlags(fs)
		fs.Usage = usageFunc(wantCmd, cmd, fs)

		if help {
			fs.Usage()
			os.Exit(0)
		}

		_ = fs.Parse(gfs.Args()[1:])

		cmd.handle(gf, fs)

		os.Exit(0)
	}

	// Command not found.
	flog.Error("command %q not found", wantCmd)
	gfs.Usage()
	os.Exit(2)
}

func usageFunc(wantCmd string, cmd command, flagSet *flag.FlagSet) func() {
	return func() {
		help := fmt.Sprintf(`NAME:
	%s %s - %s

USAGE:
	%s %s %s

DESCRIPTION:
	%s`,
			os.Args[0], wantCmd, cmd.spec().shortDesc,
			os.Args[0], wantCmd, cmd.spec().usage,
			cmd.spec().longDesc,
		)

		flagHelpMsg := flagHelp(flagSet)
		if flagHelpMsg == "" {
			fmt.Printf("%s", help)
		} else {
			fmt.Printf("%s\n\n%s", help, flagHelpMsg)
		}
	}
}

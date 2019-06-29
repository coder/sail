package main

import (
	"flag"
	"fmt"
	"unicode/utf8"

	"github.com/posener/complete"

	"go.coder.com/cli"
)

func genAutocomplete(cmds []cli.Command) complete.Command {
	var (
		ac = complete.Command{
			Sub:         map[string]complete.Command{},
			Flags:       map[string]complete.Predictor{},
			GlobalFlags: map[string]complete.Predictor{},
		}
	)

	// add all commands + flags
	for _, e := range cmds {
		// the root command is handled separately and its flags are added as global flags.
		if e.Spec().Name == "sail" {
			registerFlags(e.(cli.FlaggedCommand), func(f *flag.Flag) {
				n := fmtFlag(f.Name)
				switch f.Name {
				// special case for autocompleting configs
				case "config":
					ac.GlobalFlags[n] = complete.PredictFiles("*.toml")
				default:
					ac.GlobalFlags[n] = complete.PredictAnything
				}
			})

			// don't register root command
			continue
		}

		genCommandAutocomplete(ac, e)
	}

	return ac
}

// genCommandAutocomplete generates an autocomplete entry for a command.
// It will recursively add all subcommands.
func genCommandAutocomplete(parent complete.Command, cmd cli.Command) {
	child := complete.Command{
		Sub:   map[string]complete.Command{},
		Flags: map[string]complete.Predictor{},
	}

	if f, ok := cmd.(cli.FlaggedCommand); ok {
		registerFlags(f, func(f *flag.Flag) {
			// TODO: we can probably write a wrapper around *flag.FlagSet
			// that is smarter about predictions
			child.Flags[fmtFlag(f.Name)] = complete.PredictAnything
		})
	}

	if pc, ok := cmd.(cli.ParentCommand); ok {
		genSubcommandAutocomplete(child, pc.Subcommands())
	}

	parent.Sub[cmd.Spec().Name] = child
	return
}

// genSubcommands recursively walks up a command tree, adding child commands to their parent.
func genSubcommandAutocomplete(parent complete.Command, cmds []cli.Command) {
	for _, e := range cmds {
		genCommandAutocomplete(parent, e)
	}
}

func registerFlags(cmd cli.FlaggedCommand, visitFunc func(f *flag.Flag)) {
	// make a fake FlagSet for the command to set the flags on,
	// then we can iterate over them
	set := flag.NewFlagSet("", flag.ContinueOnError)
	cmd.RegisterFlags(set)

	set.VisitAll(visitFunc)
}

func fmtFlag(name string) string {
	if utf8.RuneCountInString(name) > 1 {
		return fmt.Sprintf("--%s", name)

	}

	return fmt.Sprintf("-%s", name)
}

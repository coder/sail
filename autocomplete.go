package main

import (
	"flag"
	"fmt"

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
	for i, e := range cmds {
		// root is always first, handle these as global flags
		if i == 0 {
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
		}

		cmd := complete.Command{
			Flags: map[string]complete.Predictor{},
		}

		if f, ok := e.(cli.FlaggedCommand); ok {
			registerFlags(f, func(f *flag.Flag) {
				// TODO: we can probably write a wrapper around *flag.FlatSet
				// that is smarter about predictions
				cmd.Flags[fmtFlag(f.Name)] = complete.PredictAnything
			})
		}

		spec := e.Spec()
		ac.Sub[spec.Name] = cmd
	}

	return ac
}

func registerFlags(cmd cli.FlaggedCommand, visitFunc func(f *flag.Flag)) {
	// make a fake flag set so the command will set the flags,
	// then iterate over them
	set := flag.NewFlagSet("", flag.ContinueOnError)
	cmd.RegisterFlags(set)

	set.VisitAll(visitFunc)
}

func fmtFlag(name string) string {
	if len(name) == 1 {
		return fmt.Sprintf("-%s", name)

	}

	return fmt.Sprintf("--%s", name)
}

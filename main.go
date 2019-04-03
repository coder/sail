package main

import (
	"flag"
	"fmt"
	"strings"
)

func main() {
	flg := initFlags()

	flag.Usage = func() {
		var flagHelp strings.Builder
		flag.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(&flagHelp, "\t%v\t%v\t(%v)\n", f.Name, f.Usage, f.DefValue)
		})

		fmt.Printf(`Usage: narwhal <repo>

narwhal is a utility for managing Docker-based, VS Code environments.
More info: https://github.com/codercom/narwhal

Arguments:
	repo	A Git repo. If you're using GitHub, just "org/reponame" is necessary.
		Use the full SSH clone address otherwise.

Flags:
%v
`, flagHelp.String())
	}
	flag.Parse()

	c := mustReadConfig(flg.configPath)

	flg.run(c)
}

package main

import (
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"go.coder.com/flog"
	"go.coder.com/narwhal"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// config describes the config.toml.
// Changes to this
type config struct {
	DefaultHost string `toml:"default_host"`
}

func readConfig(path string) config {
	var c config
	_, err := toml.DecodeFile(path, &c)
	if err != nil {
		if os.IsNotExist(err) {
			flog.Info("No configuration exists at %v, writing default.", path)

			baseDir := filepath.Dir(path)
			err = os.MkdirAll(baseDir, 0755)
			if err != nil {
				flog.Fatal("failed to mkdirall %v: %v", baseDir, err)
			}

			err = ioutil.WriteFile(path, []byte(narwhal.DefaultConfig), 0644)
			if err != nil {
				flog.Fatal("failed to write default config @ %v\n%v", path, err)
			}

			return readConfig(path)
		}
		flog.Fatal("failed to parse config @ %v\n%v", path, err)
	}
	return c
}

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	verbose := flag.Bool("v", false, "Enable debug logging.")
	configPath := flag.String("config", homeDir+"/.config/narwhal/narwhal.toml", "Path to config.")

	flag.Usage = func() {
		var flagHelp strings.Builder
		flag.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(&flagHelp, "\t%v\t%v\t(%v)\n", f.Name, f.Usage, f.DefValue)
		})

		fmt.Printf(`Usage: narwhal <[user@]host> <repo>

narwhal is a utility for managing remote, Docker-based, code-server environments.
More info: https://github.com/codercom/narwhal

Arguments:
	host	An SSH server, uses local user as username if none provided. The 
		special host "local" uses the local system.
	repo	A Git repo. If you're using GitHub, just "org/reponame" is necessary.
		Use the full SSH clone address otherwise.

Flags:
%v
`, flagHelp.String())
	}
	flag.Parse()

	_ = verbose

	c := readConfig(*configPath)

	flog.Info("config: %+v", c)

	//
	//host := flag.Arg(0)
	//repo := flag.Arg(1)
	//
	//if host == "" {
	//
	//}
}

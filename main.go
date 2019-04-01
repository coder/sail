package main

import (
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/fatih/color"
	"go.coder.com/flog"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func readConfig(path string) Config {
	var c Config
	_, err := toml.DecodeFile(path, &c)
	if err != nil {
		if os.IsNotExist(err) {
			flog.Info("No configuration exists at %v, writing default.", path)

			baseDir := filepath.Dir(path)
			err = os.MkdirAll(baseDir, 0755)
			if err != nil {
				flog.Fatal("failed to mkdirall %v: %v", baseDir, err)
			}

			err = ioutil.WriteFile(path, []byte(DefaultConfig), 0644)
			if err != nil {
				flog.Fatal("failed to write default config @ %v\n%v", path, err)
			}

			return readConfig(path)
		}
		flog.Fatal("failed to parse config @ %v\n%v", path, err)
	}
	return c
}

type flags struct {
	verbose    bool
	configPath string
	newContainer bool
}

func (flg *flags) debug(msg string, args ...interface{}) {
	if !flg.verbose {
		return
	}
	flog.Log(
		flog.Level(color.New(color.FgHiMagenta).Sprint("DEBUG")),
		msg, args...,
	)
}

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	var flg flags

	flag.BoolVar(&flg.verbose, "v", false, "Enable debug logging.")
	flag.StringVar(&flg.configPath, "config", homeDir+"/.config/narwhal/narwhal.toml", "Path to config.")
	flag.BoolVar(&flg.newContainer, "new", false, "Force create a new container.")

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

	c := readConfig(flg.configPath)
	run(flg, c)
}

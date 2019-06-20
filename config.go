package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"go.coder.com/flog"
)

func resolvePath(homedir string, path string) string {
	path = filepath.Clean(path)

	// So homedir resolution is possible in absolute paths.
	if filepath.IsAbs(path) {
		return path
	}

	// Replace tilde notation in path with homedir.
	if path == "~" {
		path = homedir
	} else if strings.HasPrefix(path, "~/") {
		path = filepath.Join(homedir, path[2:])
	}

	return filepath.Clean(path)
}

// config describes the config.toml.
// Changes to this should be accompanied by changes to DefaultConfig.
type config struct {
	DefaultImage        string `toml:"default_image"`
	ProjectRoot         string `toml:"project_root"`
	DefaultHat          string `toml:"default_hat"`
	DefaultSchema       string `toml:"default_schema"`
	DefaultHost         string `toml:"default_host"`
	DefaultOrganization string `toml:"default_organization"`
}

// DefaultConfig is the default configuration file string.
const DefaultConfig = `# sail configuration.
# default_image is the default Docker image to use if the repository provides none.
default_image = "codercom/ubuntu-dev"

# project_root is the base from which projects are mounted.
# projects are stored in directories with form "<root>/<org>/<repo>"
project_root = "~/Projects"

# default hat lets you configure a hat that's applied automatically by default.
# default_hat = ""

# default schema used to clone repo in sail run if none given
default_schema = "ssh"

# default host used to clone repo in sail run if none given
default_host = "github.com"

# default_oranization lets you configure which username to use on default_host
# when cloning a repo.
# default_organization = ""
`

// metaRoot returns the root path of all metadata stored on the host.
func metaRoot() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	return filepath.Join(homeDir, ".config", "sail")
}

func mustReadConfig(path string) config {
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

			err = ioutil.WriteFile(path, []byte(DefaultConfig), 0644)
			if err != nil {
				flog.Fatal("failed to write default config @ %v\n%v", path, err)
			}

			return mustReadConfig(path)
		}
		flog.Fatal("failed to parse config @ %v\n%v", path, err)
	}
	return c
}

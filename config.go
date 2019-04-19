package main

import (
	"fmt"
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

	list := strings.Split(path, string(filepath.Separator))

	for i, seg := range list {
		if seg == "~" {
			list[i] = homedir
		}
	}

	return filepath.Join(list...)
}

// config describes the config.toml.
// Changes to this should be accompanied by changes to DefaultConfig.
type config struct {
	DefaultImage   string `toml:"default_image"`
	ProjectRoot    string `toml:"project_root"`
	DefaultHat     string `toml:"default_hat"`
	DefaultNetwork string `toml:"default_network"`
	DefaultSubnet  string `toml:"default_subnet"`
}

func (c config) setEmptyToDefault() config {
	if c.DefaultImage == "" {
		c.DefaultImage = "codercom/ubuntu-dev"
	}
	if c.ProjectRoot == "" {
		c.ProjectRoot = "~/Projects"
	}
	if c.DefaultNetwork == "" {
		c.DefaultNetwork = "sail"
	}
	if c.DefaultSubnet == "" {
		c.DefaultSubnet = "172.20.0.0/16"
	}

	return c
}

func (c config) tomlString() string {
	return fmt.Sprintf(`# sail configuration.
# default_image is the default Docker image to use if the repository provides none.
default_image = "%s"

# project_root is the base from which projects are mounted.
# projects are stored in directories with form "<root>/<org>/<repo."
project_root = "%s"

# default hat lets you configure a hat that's applied automatically by default.
# default_hat = "%s"

# default_network is the name of the docker network to use for creating and
# configuring containers.
default_network = "%s"

# default_subnet is the subnet to use if we need to create the default network.
# If the default network already exists with a different subnet, the existing
# subnet will be used.
default_subnet = "%s"

`, c.DefaultImage, c.ProjectRoot, c.DefaultHat, c.DefaultNetwork, c.DefaultSubnet)
}

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

			c = c.setEmptyToDefault()
			err = ioutil.WriteFile(path, []byte(c.tomlString()), 0644)
			if err != nil {
				flog.Fatal("failed to write default config @ %v\n%v", path, err)
			}

			return mustReadConfig(path)
		}
		flog.Fatal("failed to parse config @ %v\n%v", path, err)
	}
	return c.setEmptyToDefault()
}

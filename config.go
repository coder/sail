package main

import (
	"github.com/BurntSushi/toml"
	"go.coder.com/flog"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func resolvePath(homedir string, path string) string {
	path = filepath.Clean(path)

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
	DefaultImage         string            `toml:"default_image"`
	ProjectRoot          string            `toml:"project_root"`
	DefaultHat           string            `toml:"default_hat"`
	Shares               map[string]string `toml:"shares"`
}

const DefaultConfig = `# Narwhal configuration.
# default_image is the default Docker image to use if the repository provides none.
default_image = "codercom/ubuntu-dev"

# project_root is the base from which projects are mounted.
# projects are stored in directories with form "<root>/<org>/<repo."
project_root = "~/Projects"

# default hat lets you configure a hat that's applied automatically by default.
# default_hat = ""


[shares]
# These shares synchronizes VS Code settings.
"~/.config/Code" = "~/.config/Code"
"~/.vscode/extensions" = "~/.vscode/extensions"

# Things you probably want inside.
"~/.gitconfig" = "~/.gitconfig"
"~/.ssh" = "~/.ssh"
`

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

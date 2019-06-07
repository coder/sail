package main

import (
	"os"
	"path/filepath"
	"runtime"
)

const (
	vsCodeConfigDirEnv     = "VSCODE_CONFIG_DIR"
	vsCodeExtensionsDirEnv = "VSCODE_EXTENSIONS_DIR"
)

func vscodeConfigDir() string {
	if env, ok := os.LookupEnv(vsCodeConfigDirEnv); ok {
		return os.ExpandEnv(env)
	}

	path := os.ExpandEnv("$HOME/.config/Code/")
	if runtime.GOOS == "darwin" {
		path = os.ExpandEnv("$HOME/Library/Application Support/Code/")
	}
	return filepath.Clean(path)
}

func vscodeExtensionsDir() string {
	if env, ok := os.LookupEnv(vsCodeExtensionsDirEnv); ok {
		return os.ExpandEnv(env)
	}

	path := os.ExpandEnv("$HOME/.vscode/extensions/")
	return filepath.Clean(path)
}

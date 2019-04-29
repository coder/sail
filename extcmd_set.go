package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"go.coder.com/cli"
	"go.coder.com/flog"
	"go.coder.com/sail/internal/extensions"
	"golang.org/x/xerrors"
)

type extSetCmd struct{}

func (c *extSetCmd) Spec() cli.CommandSpec {
	return cli.CommandSpec{
		Name:  "set",
		Usage: "set [Dockerfile]",
		Desc: `Appends or replaces an existing RUN statement that adds the locally installed extensions to the provided Dockerfile, or creates a new one if it doesn't exist.
If no Dockerfile is given, it will be sent to stdout.

Example:
sail extensions set Dockerfile
		`,
		Hidden: false,
	}
}

func (c *extSetCmd) Run(fl *flag.FlagSet) {
	err := c.set(fl)
	if err != nil {
		flog.Error("failed to set extension list: %s", err)
	}
}

// set overrides all previously set extensions with the ones in the local VS Code extension folder.
func (c *extSetCmd) set(fl *flag.FlagSet) error {
	var (
		file       = fl.Arg(0)
		fileExists = false
	)

	stat, err := os.Stat(file)
	if err == nil {
		fileExists = true
	}

	extDir, err := vscodeExtensionDir()
	if err != nil {
		return xerrors.Errorf("failed to find vscode edition: %w", err)
	}

	exts, err := extensions.ParseExtensionList(extDir)
	if err != nil {
		return xerrors.Errorf("failed to parse extension list: %w", err)
	}

	if len(exts) == 0 {
		fmt.Println("No extensions found.")
		// is adding a config option the best way for this?
		fmt.Println("Is this a mistake? Edit your sail.toml with the correct VS Code path.")
		return nil
	}

	if fileExists {
		raw, err := ioutil.ReadFile(file)
		if err != nil {
			return xerrors.Errorf("failed to read Dockerfile: %w", err)
		}

		r, err := extensions.DockerfileSetExtensions(raw, exts)
		// we only care if this executes successfully beacuse if it errors,
		// that means we can't find the extension block and it'll be added
		// below.
		if err == nil {
			err = ioutil.WriteFile(file, r, stat.Mode())
			if err != nil {
				return xerrors.Errorf("failed to write Dockerfile: %w", err)
			}

			return nil
		}

	}

	buf := bytes.NewBuffer(nil)

	// file doesnt exist, but we want to write to one, so format the whole Dockerfile.
	if !fileExists && file != "" {
		r, err := extensions.DockerfileSetExtensions(nil, exts)
		if err != nil {
			return xerrors.Errorf("failed to set extensions: %w", err)
		}
		buf.Write(r)

		// all other situations we just want the RUN statement
	} else {
		buf.Write(joinNewline(extensions.FmtExtensions(exts)))
	}

	// we're not writing to a file, so copy to terminal
	if file == "" {
		_, err := io.Copy(os.Stdout, buf)
		if err != nil {
			return xerrors.Errorf("failed to copy Dockerfile to stdout: %w", err)
		}

		return nil
	}

	fi, err := os.OpenFile(file, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return xerrors.Errorf("failed to open Dockerfile: %w", err)
	}
	defer func() {
		err := fi.Close()
		if err != nil {
			flog.Error("failed to close Dockerfile: %s", err)
		}
	}()

	_, err = io.Copy(fi, buf)
	if err != nil {
		return xerrors.Errorf("failed to write to Dockerfile: %w", err)
	}

	return nil
}

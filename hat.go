package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"go.coder.com/flog"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// dockerReplaceFrom replaces the FROM clause in a Dockerfile
// with the provided base.
func dockerReplaceFrom(dockerFile []byte, base string) []byte {
	buf := bytes.NewBuffer(make([]byte, 0, len(dockerFile)))

	sc := bufio.NewScanner(bytes.NewReader(dockerFile))

	for sc.Scan() {
		byt := sc.Bytes()
		if bytes.HasPrefix(byt, []byte("FROM")) {
			byt = []byte("FROM " + base)
		}

		buf.Write(byt)
		buf.WriteByte('\n')
	}
	return bytes.TrimSpace(buf.Bytes())
}

// resolveGitHubHat takes a path like ammario/dotfiles
// and downloads it into a temporary direcory.
func resolveGitHubHat(ghPath string) string {
	dir, err := ioutil.TempDir("", "hat")
	if err != nil {
		flog.Fatal("failed to create tempdir: %v", err)
	}

	cmd := fmtExec("git clone git@github.com:%v.git %v", ghPath, dir)
	attach(cmd)
	err = cmd.Run()
	if err != nil {
		flog.Fatal("failed to clone hat: %v", err)
	}

	return dir
}

// applyHat applies a hat, if configured, to image.
func applyHat(flg flags, c config, image string) string {
	var hatPath string
	switch {
	case flg.hat != "":
		hatPath = flg.hat
	case c.DefaultHat != "":
		hatPath = c.DefaultHat
	case flg.hat == "" && c.DefaultHat == "":
		return image
	}

	const ghPrefix = "github:"

	if strings.HasPrefix(hatPath, ghPrefix) {
		hatPath = strings.TrimLeft(hatPath, ghPrefix)
		hatPath = resolveGitHubHat(hatPath)
	}

	dockerFilePath := filepath.Join(hatPath, "Dockerfile")

	dockerFileByt, err := ioutil.ReadFile(dockerFilePath)
	if err != nil {
		flog.Fatal("failed to read %v: %v", dockerFilePath, err)
	}
	dockerFileByt = dockerReplaceFrom(dockerFileByt, image)

	fi, err := ioutil.TempFile("", "hat")
	if err != nil {
		flog.Fatal("failed to create temp file: %v", err)
	}
	defer fi.Close()
	defer os.Remove(fi.Name())

	_, err = fi.Write(dockerFileByt)
	if err != nil {
		flog.Fatal("failed to write to %v: %v", fi.Name(), err)
	}

	// We tag based on the checksum of the Dockerfile to avoid spamming
	// images.
	csm := sha256.Sum256(dockerFileByt)
	imageName := image + "-hat-" + hex.EncodeToString(csm[:])[:16]

	flog.Info("building hat image %v", imageName)
	cmd := fmtExec("docker build -t %v -f %v %v",
		imageName, fi.Name(), hatPath,
	)
	attach(cmd)
	err = cmd.Run()
	if err != nil {
		flog.Fatal("failed to build hatted image: %v", err)
	}
	return imageName
}

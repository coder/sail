package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"golang.org/x/xerrors"

	"go.coder.com/flog"
)

type globalFlags struct {
	verbose    bool
	configPath string
}

func (gf *globalFlags) debug(msg string, args ...interface{}) {
	if !gf.verbose {
		return
	}

	flog.Log(
		flog.Level(color.New(color.FgHiMagenta).Sprint("DEBUG")),
		msg, args...,
	)
}

func (gf *globalFlags) config() config {
	return mustReadConfig(gf.configPath)
}

// ensureDockerDaemon verifies that Docker is running.
func (gf *globalFlags) ensureDockerDaemon() {
	if runtime.GOOS == "darwin" {
		path := os.Getenv("PATH")
		localBin := "/usr/local/bin"
		if !strings.Contains(path, localBin) {
			sep := fmt.Sprintf("%c", os.PathListSeparator)
			// Fix for MacOS to include /usr/local/bin where docker is commonly installed which is not included in $PATH when sail is launched by browser that was opened in Finder
			os.Setenv("PATH", strings.Join([]string{path, localBin}, sep))
		}
	}
	out, err := exec.Command("docker", "info").CombinedOutput()
	if err != nil {
		flog.Fatal("failed to run `docker info`: %v\n%s", err, out)
	}
	gf.debug("verified Docker is running")
}

func requireRepo(conf config, prefs schemaPrefs, fl *flag.FlagSet) repo {
	var (
		repoURI = strings.Join(fl.Args(), "/")
		r       repo
		err     error
	)

	if repoURI == "" {
		flog.Fatal("Argument <repo> must be provided.")
	}

	// if this returns a non-empty string know it's pointing to a valid project on disk
	// an error indicates an existing path outside of the project dir
	repoName, err := pathIsRunnable(conf, repoURI)
	if err != nil {
		flog.Fatal(err.Error())
	}

	if repoName != "" {
		// we only need the path since the repo exists on disk.
		// there's not currently way for us to figure out the host anyways
		r = repo{URL: &url.URL{Path: repoName}}
	} else {
		r, err = parseRepo(defaultSchema(conf, prefs), conf.DefaultHost, conf.DefaultOrganization, repoURI)
		if err != nil {
			flog.Fatal("failed to parse repo %q: %v", repoURI, err)
		}
	}

	// check if path is pointing to a subdirectory
	if sp := strings.Split(r.Path, "/"); len(sp) > 2 {
		r.Path = strings.Join(sp[:2], "/")
		r.subdir = strings.Join(sp[2:], "/")
	}

	return r
}

// pathIsRunnable returns the container name if the given path exists and is
// in the projects directory, else an empty string. An error is returned if
// and only if the path exists but it isn't in the user's project directory.
func pathIsRunnable(conf config, path string) (cnt string, _ error) {
	fp, err := filepath.Abs(path)
	if err != nil {
		return
	}

	s, err := os.Stat(fp)
	if err != nil {
		return
	}

	if !s.IsDir() {
		return
	}

	pre := expandRoot(conf.ProjectRoot)
	if pre[len(pre)-1] != '/' {
		pre = pre + "/"
	}

	// path exists but doesn't belong to projects directory, return error
	if !strings.HasPrefix(fp, pre[:len(pre)-1]) {
		return "", xerrors.Errorf("directory %s exists but isn't in projects directory", fp)
	}

	split := strings.Split(fp, "/")
	if len(split) < 2 {
		return
	}

	return strings.TrimPrefix(fp, pre), nil
}

func expandRoot(path string) string {
	u, _ := user.Current()
	return strings.Replace(path, "~/", u.HomeDir+"/", 1)
}

func defaultSchema(conf config, prefs schemaPrefs) string {
	switch {
	case prefs.ssh:
		return "ssh"
	case prefs.https:
		return "https"
	case prefs.http:
		return "http"
	case conf.DefaultSchema != "":
		return conf.DefaultSchema
	default:
		return "ssh"
	}
}

// project reads the project as the first parameter.
func (gf *globalFlags) project(prefs schemaPrefs, fl *flag.FlagSet) *project {
	conf := gf.config()
	return &project{
		conf: conf,
		repo: requireRepo(conf, prefs, fl),
	}
}

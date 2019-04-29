package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path"
	"regexp"
	"strings"

	"github.com/google/go-github/v24/github"
	"go.coder.com/flog"
)

var repoRegex = regexp.MustCompile(`^(?P<user>\w+@)?(?P<host>\S+:)?(?P<path>\S+)$`)

// subexpMap returns a map with keys of form subexpName -> value.
func subexpMap(r *regexp.Regexp, target string) map[string]string {
	matches := r.FindStringSubmatch(target)
	m := make(map[string]string, len(r.SubexpNames()))
	for i, name := range r.SubexpNames() {
		if i > len(matches) { // no more matches
			return m
		}
		if i == 0 { // first subexp is the target
			continue
		}
		m[name] = matches[i]
	}
	return m
}

type repo struct {
	User, Host, Path string
}

func (r repo) CloneURI() string {
	return fmt.Sprintf("%v@%v:%v", r.User, r.Host, r.Path)
}

func (r repo) DockerName() string {
	return toDockerName(
		r.trimPathGitSuffix(),
	)
}

func (r repo) BaseName() string {
	return strings.TrimSuffix(path.Base(r.Path), ".git")
}

// ParseRepo parses a reponame into a repo.
// The default user is Git.
// The default Host is github.com.
// If the host is github.com, `.git` is always at the end of Path.
func ParseRepo(name string) (repo, error) {
	m := subexpMap(repoRegex, name)

	r := repo{
		User: strings.TrimSuffix(m["user"], "@"),
		Host: strings.TrimSuffix(m["host"], ":"),
		Path: m["path"],
	}

	if r.User == "" {
		r.User = "git"
	}

	if r.Host == "" {
		r.Host = "github.com"
	}

	if r.Path == "" {
		return repo{}, errors.New("no path provided")
	}

	if r.Host == "github.com" && !strings.HasSuffix(r.Path, ".git") {
		r.Path += ".git"
	}

	if r.Host == "github.com" && !strings.Contains(r.Path, "/") {
		return repo{}, errors.New("GitHub repo path must have /")
	}

	return r, nil
}

func (r repo) trimPathGitSuffix() string {
	return strings.TrimSuffix(r.Path, ".git")
}

// language returns the language of a repository using github's detected language.
// This is a best effort try and will return the empty string if something fails.
func (r repo) language() string {
	orgRepo := strings.SplitN(r.trimPathGitSuffix(), "/", 2)
	if len(orgRepo) != 2 {
		return ""
	}

	repo, resp, err := github.NewClient(nil).Repositories.Get(context.Background(), orgRepo[0], orgRepo[1])
	if err != nil {
		flog.Error("unable to get repo language: %v", err)
		return ""
	}

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	return repo.GetLanguage()
}

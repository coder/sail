package main

import (
	"context"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/google/go-github/v24/github"
	"golang.org/x/xerrors"

	"go.coder.com/flog"
)

type Repo struct {
	*url.URL
}

func (r Repo) CloneURI() string {
	return r.String()
}

func (r Repo) DockerName() string {
	return toDockerName(
		r.trimPath(),
	)
}

func (r Repo) trimPath() string {
	return strings.TrimPrefix(r.Path, "/")
}

func (r Repo) BaseName() string {
	return strings.TrimSuffix(path.Base(r.Path), ".git")
}

// ParseRepo parses a reponame into a repo.
// The default user is Git.
// The default Host is github.com.
// If the host is github.com, `.git` is always at the end of Path.
func ParseRepo(defaultSchema, name string) (Repo, error) {
	u, err := url.Parse(name)
	if err != nil {
		return Repo{}, xerrors.Errorf("failed to parse repo path: %w", err)
	}

	r := Repo{u}

	if r.Scheme == "" {
		r.Scheme = defaultSchema
	}

	// this probably means the host is part of the path
	if r.Host == "" {
		parts := strings.Split(r.trimPath(), "/")
		// we would expect there to be 3 parts if the host is part of the path
		if len(parts) == 3 {
			r.Host = parts[0]
			r.Path = strings.Join(parts[1:], "/")
		} else {
			// as a default case we assume github
			r.Host = "github.com"
		}
	}

	// make sure path doesn't have a leading forward slash
	r.Path = strings.TrimPrefix(r.Path, "/")

	// non-existent or invalid path
	if r.Path == "" || len(strings.Split(r.Path, "/")) != 2 {
		return Repo{}, xerrors.Errorf("invalid repo: %s", r.Path)
	}

	// if host contains a username, e.g. git@github.com
	// url.Parse accidentally leaves this in the host if there isnt a schema in front of it
	if sp := strings.Split(r.Host, "@"); len(sp) > 1 {
		// split username/password if exists
		usp := strings.Split(sp[0], ":")
		switch len(usp) {
		case 1:
			r.User = url.User(usp[0])
		case 2:
			r.User = url.UserPassword(usp[0], usp[1])
		default:
			return Repo{}, xerrors.Errorf("invalid user: %s", sp[0])
		}

		// remove username from host
		r.Host = strings.Join(sp[1:], "@")
	}

	// don't set git as username for http urls
	if r.User.Username() == "" && r.Scheme == "ssh" {
		r.User = url.User("git")
	}

	return r, nil
}

// language returns the language of a repository using github's detected language.
// This is a best effort try and will return the empty string if something fails.
func (r Repo) language() string {
	orgRepo := strings.SplitN(r.trimPath(), "/", 2)
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

func isAllowedSchema(s string) bool {
	return s == "http" ||
		s == "https" ||
		s == "ssh"
}

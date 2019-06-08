package environment

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/google/go-github/v24/github"
	"go.coder.com/flog"
	"golang.org/x/xerrors"
)

type Repo struct {
	*url.URL
}

// ParseRepo parses a Reponame into a Repo.
//
// It can be a full url like https://github.com/cdr/sail or
// ssh://git@github.com/cdr/sail, or just the path like cdr/sail and the host +
// schema will be inferred.  By default the host and the schema will be the
// provided defaultSchema.
func ParseRepo(defaultSchema, defaultHost, name string) (Repo, error) {
	u, err := url.Parse(name)
	if err != nil {
		return Repo{}, xerrors.Errorf("failed to parse Repo path: %w", err)
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
			r.Host = defaultHost
		}
	}

	// make sure path doesn't have a leading forward slash
	r.Path = strings.TrimPrefix(r.Path, "/")

	// make sure the path doesn't have a trailing .git
	r.Path = strings.TrimSuffix(r.Path, ".git")

	// non-existent or invalid path
	if r.Path == "" || len(strings.Split(r.Path, "/")) != 2 {
		return Repo{}, xerrors.Errorf("invalid Repo: %s", r.Path)
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

func (r Repo) CloneURI() string {
	uri := r.String()
	if !strings.HasSuffix(uri, ".git") {
		return fmt.Sprintf("%s.git", uri)
	}
	return uri
}

func (r Repo) DockerName() string {
	return strings.Replace(r.trimPath(), "/", "_", 1)
}

func (r Repo) trimPath() string {
	return strings.TrimPrefix(r.Path, "/")
}

func (r Repo) BaseName() string {
	return strings.TrimSuffix(path.Base(r.Path), ".git")
}

// Language returns the language of a Repository using github's detected
// language.  This is a best effort try and will return the empty string if
// something fails.
func (r Repo) Language() string {
	orgRepo := strings.SplitN(r.trimPath(), "/", 2)
	if len(orgRepo) != 2 {
		return ""
	}

	Repo, resp, err := github.NewClient(nil).Repositories.Get(context.Background(), orgRepo[0], orgRepo[1])
	if err != nil {
		flog.Error("unable to get Repo language: %v", err)
		return ""
	}

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	return strings.ToLower(Repo.GetLanguage())
}

func isAllowedSchema(s string) bool {
	return s == "http" ||
		s == "https" ||
		s == "ssh"
}

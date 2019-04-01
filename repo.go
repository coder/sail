package main

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var repoRegex = regexp.MustCompile(`^(?P<user>\w+@)?(?P<host>\S+:)?(?P<path>\S+)$`)

// subexpMap returns a map with keys of form subexpName -> value.
func subexpMap(r *regexp.Regexp, target string) map[string]string {
	matches := r.FindStringSubmatch(target)
	m := make(map[string]string, len(r.SubexpNames()))
	for i, name := range r.SubexpNames() {
		if i > len(matches) { //no more matches
			return m
		}
		if i == 0 { //first subexp is the target
			continue
		}
		m[name] = matches[i]
	}
	return m
}

type Repo struct {
	User, Host, Path string
}

func (r Repo) String() string {
	return fmt.Sprintf("%v@%v:%v", r.User, r.Host, r.Path)
}

func (r Repo) DockerName() string {
	return strings.Replace(
		strings.TrimSuffix(r.Path, ".git"), "/", "-", -1,
	)
}

// ParseRepo parses a reponame into a Repo.
// The default user is Git.
// The default Host is github.com.
// If the host is github.com, `.git` is always at the end of Path.
func ParseRepo(name string) (Repo, error) {
	m := subexpMap(repoRegex, name)

	repo := Repo{
		User: strings.TrimSuffix(m["user"], "@"),
		Host: strings.TrimSuffix(m["host"], ":"),
		Path: m["path"],
	}

	if repo.User == "" {
		repo.User = "git"
	}

	if repo.Host == "" {
		repo.Host = "github.com"
	}

	if repo.Path == "" {
		return Repo{}, errors.New("no path provided")
	}

	if repo.Host == "github.com" && !strings.HasSuffix(repo.Path, ".git") {
		repo.Path += ".git"
	}

	return repo, nil
}

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRepo(t *testing.T) {
	var tests = []struct {
		defSchema string
		fullPath  string

		expPath     string
		expHost     string
		expUser     string
		expSchema   string
		expCloneURL string
	}{
		// ensure default schema works as expected
		{
			"ssh",
			"cdr/sail",
			"cdr/sail",
			"github.com",
			"git",
			"ssh",
			"ssh://git@github.com/cdr/sail",
		},
		// ensure default schemas works as expected
		{
			"http",
			"cdr/sail",
			"cdr/sail",
			"github.com",
			"",
			"http",
			"http://github.com/cdr/sail",
		},
		// ensure default schemas works as expected
		{
			"https",
			"cdr/sail",
			"cdr/sail",
			"github.com",
			"",
			"https",
			"https://github.com/cdr/sail",
		},
		// http url parses correctly
		{
			"https",
			"https://github.com/cdr/sail",
			"cdr/sail",
			"github.com",
			"",
			"https",
			"https://github.com/cdr/sail",
		},
		// git url with username and without schema parses correctly
		{
			"ssh",
			"git@github.com/cdr/sail.git",
			"cdr/sail.git",
			"github.com",
			"git",
			"ssh",
			"ssh://git@github.com/cdr/sail.git",
		},
		// different default schema doesn't override given schema
		{
			"http",
			"ssh://git@github.com/cdr/sail",
			"cdr/sail",
			"github.com",
			"git",
			"ssh",
			"ssh://git@github.com/cdr/sail",
		},
	}

	for _, test := range tests {
		repo, err := parseRepo(test.defSchema, test.fullPath)
		require.NoError(t, err)

		assert.Equal(t, test.expPath, repo.Path, "expected path to be the same")
		assert.Equal(t, test.expHost, repo.Host, "expected host to be the same")
		assert.Equal(t, test.expUser, repo.User.Username(), "expected user to be the same")
		assert.Equal(t, test.expSchema, repo.Scheme, "expected scheme to be the same")

		assert.Equal(t, test.expCloneURL, repo.CloneURI(), "expected clone uri to be the same")
	}
}

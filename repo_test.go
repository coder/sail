package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRepo(t *testing.T) {
	var (
		repo repo
		err  error
	)
	repo, err = ParseRepo("cdr/sail")
	require.NoError(t, err)

	assert.Equal(t, "cdr/sail.git", repo.Path)
	assert.Equal(t, "github.com", repo.Host)
	assert.Equal(t, "git", repo.User)

	repo, err = ParseRepo("git@github.com:cdr/sail.git")
	require.NoError(t, err)

	assert.Equal(t, "cdr/sail.git", repo.Path)
	assert.Equal(t, "github.com", repo.Host)
	assert.Equal(t, "git", repo.User)

	assert.Equal(t, "git@github.com:cdr/sail.git", repo.CloneURI())
}

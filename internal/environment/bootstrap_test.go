package environment

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.coder.com/sail/internal/randstr"
)

func Test_Bootstrap(t *testing.T) {
	checkDockerDaemon(t)
	t.Parallel()

	const project = "cdr/sshcode"
	repo, err := ParseRepo("https", "github.com", project)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	name := "bootstrap-test-" + randstr.MakeCharset(randstr.Lower, 5)
	cfg := NewDefaultBuildConfig(name)
	env, err := Bootstrap(ctx, cfg, &repo, "")
	require.NoError(t, err)
	defer removeEnv(t, env)

	// If git directory exists, good chance everything worked correctly.
	out, err := env.Exec(ctx, "ls", "/home/user/Projects/sshcode/.git").CombinedOutput()
	require.NoError(t, err, "failed to list git dir: %s", out)
}

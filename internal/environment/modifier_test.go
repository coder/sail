package environment

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_Modifier(t *testing.T) {
	checkDockerDaemon(t)

	t.Run("Basic", func(t *testing.T) {
		const uri = "https://github.com/cdr/sshcode"
		t.Logf("repo: %s", uri)
		r, err := ParseRepo("", "", uri)
		require.NoError(t, err)
		b := NewDefaultBuilder(&r)

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		env, err := b.Build(ctx)
		require.NoError(t, err)
		defer func() {
			err := Purge(ctx, env)
			require.NoError(t, err)
		}()

		provider := RawDockerfileProvider([]byte(`FROM codercom/ubuntu-dev
RUN sudo apt update -y && sudo apt install fish -y`))

		modifier := NewModifier(provider)
		env, err = modifier.Modify(ctx, env)
		require.NoError(t, err)

		// Simple check to see if fish was installed.
		out, err := env.exec(ctx, "fish", "-c", "echo hello").CombinedOutput()
		require.NoError(t, err, "failed to run fish: %s", out)
	})
}

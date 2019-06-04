package environment

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// checkDockerDaemon checks if the docker daemon is running. The daemon may be
// ran remotely by setting 'DOCKER_HOST'.
func checkDockerDaemon(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "docker", "info").CombinedOutput()
	if err != nil {
		t.Fatalf("failed to get docker info: %s", out)
	}

	host, exists := os.LookupEnv("DOCKER_HOST")
	if exists {
		t.Logf("docker host set to %s", host)
	} else {
		t.Logf("docker running local")
	}
}

func Test_Builder(t *testing.T) {
	checkDockerDaemon(t)

	t.Run("Basic", func(t *testing.T) {
		const uri = "https://github.com/cdr/sshcode"
		t.Logf("repo: %s", uri)
		r, err := ParseRepo("", "", uri)
		b := NewDefaultBuilder(&r)

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		env, err := b.Build(ctx)
		require.NoError(t, err)
		defer func() {
			err := Purge(ctx, env)
			require.NoError(t, err)
		}()

		// Assert that code-server was correctly started.
		// port, err := env.processPort(ctx, "code-server")
		// require.NoError(t, err)
		// t.Logf("code-server listening on port %s", port)
	})
}

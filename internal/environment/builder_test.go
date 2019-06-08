package environment

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/docker/docker/api/types/mount"
	"github.com/stretchr/testify/require"
	"go.coder.com/sail/internal/randstr"
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

func removeEnv(t *testing.T, env *Environment) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err := Purge(ctx, env)
	if err != nil {
		t.Logf("error purging environment: %v", err)
	}

	cli := dockerClient()
	defer cli.Close()

	for _, m := range env.cnt.Mounts {
		if m.Type == mount.TypeVolume {
			vol, err := findLocalVolume(ctx, m.Name)
			if isVolumeNotFoundError(err) {
				continue
			}
			require.NoError(t, err)

			err = deleteLocalVolume(ctx, vol)
			require.NoError(t, err)
		}
	}
}

func Test_Builder(t *testing.T) {
	t.Parallel()
	checkDockerDaemon(t)

	t.Run("Basic", func(t *testing.T) {
		name := "builder-test-" + randstr.MakeCharset(randstr.Lower, 5)
		t.Logf("create env: %s", name)
		cfg := NewDefaultBuildConfig(name)

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		env, err := Build(ctx, cfg)
		require.NoError(t, err)
		defer removeEnv(t, env)

		// Assert that code-server was correctly started.
		// port, err := env.processPort(ctx, "code-server")
		// require.NoError(t, err)
		// t.Logf("code-server listening on port %s", port)
	})
}

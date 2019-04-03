package main

import (
	"context"
	"github.com/docker/docker/client"
	"go.coder.com/flog"
	"strings"
)

// populateImageShares adds a list of shares to the shares map from the image.
func populateImageShares(image string, shares map[string]string) {
	cli, err := client.NewEnvClient()
	if err != nil {
		flog.Fatal("failed to create docker client: %v", err)
	}

	ins, _, err := cli.ImageInspectWithRaw(context.Background(), image)
	if err != nil {
		flog.Fatal("failed to inspect %v: %v", image, err)
	}
	for k, v := range ins.ContainerConfig.Labels {
		const prefix = "share."
		if !strings.HasPrefix(k, prefix) {
			continue
		}

		tokens := strings.Split(v, ":")
		if len(tokens) != 2 {
			flog.Fatal("invalid share %q", v)
		}
		shares[tokens[0]] = tokens[1]
	}

}

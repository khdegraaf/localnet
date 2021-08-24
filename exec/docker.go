package exec

import (
	"context"
	"os/exec"
)

var docker string

func init() {
	for _, app := range []string{"podman", "docker"} {
		if _, err := exec.LookPath(app); err == nil {
			docker = app
			break
		}
	}
}

// Docker runs docker command
func Docker(ctx context.Context, args ...string) error {
	return Run(ctx, exec.Command(docker, args...))
}

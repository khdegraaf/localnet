package exec

import (
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
func Docker(args ...string) *exec.Cmd {
	return exec.Command(docker, args...)
}

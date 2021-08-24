package exec

import (
	"context"
	"os/exec"
)

var tmate string

func init() {
	for _, app := range []string{"tmate", "tmux"} {
		if _, err := exec.LookPath(app); err == nil {
			tmate = app
			break
		}
	}
}

// Tmate runs tmate command
func Tmate(ctx context.Context, args ...string) error {
	return Run(ctx, exec.Command(tmate, args...))
}

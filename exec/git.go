package exec

import (
	"context"
	"os/exec"
)

// Git runs git command
func Git(ctx context.Context, args ...string) error {
	return Run(ctx, exec.Command("git", args...))
}

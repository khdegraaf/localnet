package exec

import (
	"context"
	"io"
	"os"
	"os/exec"
)

var tmux string

func init() {
	// FIXME (wojciech): try to figure out why tmate can't attach to session with error "no current session"
	for _, app := range []string{"tmux"} {
		if _, err := exec.LookPath(app); err == nil {
			tmux = app
			break
		}
	}
}

// TMux runs tmux command
func TMux(ctx context.Context, args ...string) error {
	return Run(ctx, exec.Command(tmux, args...))
}

// TMuxNoOut runs tmux command with discarded outputs
func TMuxNoOut(ctx context.Context, args ...string) error {
	cmd := exec.Command(tmux, args...)
	cmd.Stderr = io.Discard
	cmd.Stdout = io.Discard
	return Run(ctx, cmd)
}

// TMuxTty runs tmux command with terminal attached
func TMuxTty(ctx context.Context, args ...string) error {
	// For some reason /dev/stdin, /dev/stdout and /dev/stderr point to /proc/self/fd/0 which points to /dev/null.
	// That's why tmux doesn't work if os.Stdin is assigned to cmd.Stdin.
	// Workaround is to use /proc/self/fd/1 which points to correct tty
	tty, err := os.OpenFile("/proc/self/fd/1", os.O_RDWR, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer tty.Close()

	cmd := exec.Command(tmux, args...)
	cmd.Stdin = tty
	cmd.Stderr = tty
	cmd.Stdout = tty
	return Run(ctx, cmd)
}

package exec

import (
	"context"
	"io"
	"os"
	"os/exec"
)

var tmate string

func init() {
	// FIXME (wojciech): try to figure out why tmate can't attach to session with error "no current session"
	for _, app := range []string{"tmux"} {
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

// TmateNoOut runs tmate command with discarded outputs
func TmateNoOut(ctx context.Context, args ...string) error {
	cmd := exec.Command(tmate, args...)
	cmd.Stderr = io.Discard
	cmd.Stdout = io.Discard
	return Run(ctx, cmd)
}

// TmateTty runs tmate command with terminal attached
func TmateTty(ctx context.Context, args ...string) error {
	tty, err := os.OpenFile("/dev/pts/1", os.O_RDWR, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer tty.Close()

	cmd := exec.Command(tmate, args...)
	cmd.Stdin = tty
	cmd.Stderr = tty
	cmd.Stdout = tty
	return Run(ctx, cmd)
}

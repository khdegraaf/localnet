package exec

import (
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
func TMux(args ...string) *exec.Cmd {
	return exec.Command(tmux, args...)
}

// TMuxNoOut runs tmux command with discarded outputs
func TMuxNoOut(args ...string) *exec.Cmd {
	cmd := exec.Command(tmux, args...)
	cmd.Stderr = io.Discard
	cmd.Stdout = io.Discard
	return cmd
}

// TMuxTty runs tmux command with terminal attached
func TMuxTty(tty *os.File, args ...string) *exec.Cmd {
	cmd := exec.Command(tmux, args...)
	cmd.Stdin = tty
	cmd.Stderr = tty
	cmd.Stdout = tty
	return cmd
}

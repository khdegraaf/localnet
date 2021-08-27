package tmux

import (
	"context"

	"github.com/wojciech-sif/localnet/exec"
)

const help = `
  Common keys:

    Ctrl-b 0..9       Select window 0..9
    Ctrl-b p,n        Select previous, next window
    Ctrl-b w          Select window interactively
    Ctrl-b [          Enter scrollback/copy mode (ESC to exit)
      Ctrl-s to search in scrollback in this mode
    Ctrl-b d          Detach from the terminal (keep running in background)

  Dealing with windows:

    * Exited programs will be marked in status line.

    * Ctrl-C sends SIGINT. It will stop the program inside

    * Ctrl-\ sends SIGQUIT. Use it to dump goroutines with stacktraces.

    * Press Enter to restart exited program again.
`

// NewSession returns new tmux session representation
func NewSession(name string) *Session {
	return &Session{
		name: name,
	}
}

// Session represents tmux session
type Session struct {
	name string
}

// Init initializes new tmux session if none exists
func (s *Session) Init(ctx context.Context) (bool, error) {
	if exec.Run(ctx, exec.TMuxNoOut("has-session", "-t", s.name)) == nil {
		return false, nil
	}
	return true, exec.Run(ctx, exec.TMux("new-session", "-d", "-s", s.name, "-n", "help", "bash", "-c", "trap '' SIGINT SIGQUIT; echo '"+help+"'\nwhile :; do read -sr; done"))
}

// StartApp adds application to the session
func (s *Session) StartApp(ctx context.Context, name string, args ...string) error {
	return exec.Run(ctx, exec.TMux(append([]string{"new-window", "-d", "-n", name, "-t", s.name + ":"}, args...)...))
}

// Attach attaches terminal to tmux session
func (s *Session) Attach(ctx context.Context) error {
	tty := exec.Tty()
	defer tty.Close()

	return exec.Run(ctx, exec.TMuxTty(tty, "attach-session", "-t", s.name))
}

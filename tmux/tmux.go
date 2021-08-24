package tmux

import (
	"context"
	osexec "os/exec"

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

type cmd struct {
	name string
	args []string
}

// Session represents tmux session
type Session struct {
	name     string
	commands []cmd
}

// AddWindow adds window to the session
func (s *Session) AddWindow(name string, args ...string) {
	s.commands = append(s.commands, cmd{name: name, args: args})
}

// Run attaches tmux session to terminal and opens all defined windows
func (s *Session) Run(ctx context.Context) (retErr error) {
	defer func() {
		if retErr != nil {
			return
		}
		tty := exec.Tty()
		defer tty.Close()
		retErr = exec.Run(ctx, exec.TMuxTty(tty, "attach-session", "-t", s.name))
	}()
	if exec.Run(ctx, exec.TMuxNoOut("has-session", "-t", s.name)) == nil {
		return
	}

	cmds := []*osexec.Cmd{
		exec.TMux("new-session", "-d", "-s", s.name, "-n", "help", "bash", "-c", "trap '' SIGINT SIGQUIT; echo '"+help+"'\nwhile :; do read -sr; done"),
	}
	for _, cmd := range s.commands {
		cmds = append(cmds, exec.TMux(append([]string{"new-window", "-d", "-n", cmd.name}, cmd.args...)...))
	}

	return exec.Run(ctx, cmds...)
}

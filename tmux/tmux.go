package tmux

import (
	"bytes"
	"context"
	"fmt"
	osexec "os/exec"
	"strings"

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
func NewSession(name, logDir string) *Session {
	return &Session{
		name:   name,
		logDir: logDir,
	}
}

// Session represents tmux session
type Session struct {
	name   string
	logDir string
}

// Init initializes new tmux session if none exists
func (s *Session) Init(ctx context.Context) error {
	if exec.Run(ctx, exec.TMuxNoOut("has-session", "-t", s.name)) == nil {
		return nil
	}
	return exec.Run(ctx, exec.TMux("new-session", "-d", "-s", s.name, "-n", "help", "bash", "-ce", "trap '' SIGINT SIGQUIT; echo '"+help+"'\nwhile :; do read -sr; done"))
}

// StartApp adds application to the session
func (s *Session) StartApp(ctx context.Context, name string, args ...string) error {
	return exec.Run(ctx, exec.TMux("new-window", "-d", "-n", name, "-t", s.name+":", "bash", "-ce",
		fmt.Sprintf("%s 2>&1 | tee -a \"%s/%s.log\"\nwhile :; do read -sr; done", osexec.Command("", args...).String(), s.logDir, name)))
}

// HasApp returns true if app runs in the session
func (s *Session) HasApp(ctx context.Context, name string) (bool, error) {
	buf := &bytes.Buffer{}
	cmd := exec.TMux("list-windows", "-t", s.name, "-F", "#{window_name}")
	cmd.Stdout = buf
	if err := exec.Run(ctx, cmd); err != nil {
		return false, err
	}
	for _, window := range strings.Split(buf.String(), "\n") {
		if window == name {
			return true, nil
		}
	}
	return false, nil
}

// Attach attaches terminal to tmux session
func (s *Session) Attach(ctx context.Context) error {
	tty := exec.Tty()
	defer tty.Close()

	return exec.Run(ctx, exec.TMuxTty(tty, "attach-session", "-t", s.name))
}

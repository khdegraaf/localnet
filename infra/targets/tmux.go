package targets

import (
	"context"

	"github.com/wojciech-sif/localnet/infra"
	"github.com/wojciech-sif/localnet/tmux"
)

// NewTMux creates new tmux target
func NewTMux(session *tmux.Session) *TMux {
	return &TMux{
		session: session,
	}
}

// TMux is the target deploying apps to tmux session
type TMux struct {
	session *tmux.Session
}

// DeployBinary starts binary file inside tmux session
func (t *TMux) DeployBinary(ctx context.Context, app infra.Binary) error {
	return t.session.StartApp(ctx, app.Name, append([]string{app.Path}, app.Args...)...)
}

// DeployContainer starts container inside tmux session
func (t *TMux) DeployContainer(ctx context.Context, app infra.Container) error {
	panic("not implemented yet")
}

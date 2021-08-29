package targets

import (
	"context"

	"github.com/wojciech-sif/localnet/infra"
	"github.com/wojciech-sif/localnet/tmux"
)

// NewTMux creates new tmux target
func NewTMux(config infra.Config) infra.Target {
	return &TMux{
		config: config,
		ipPool: infra.NewIPPool(config.Network),
	}
}

// TMux is the target deploying apps to tmux session
type TMux struct {
	config  infra.Config
	session *tmux.Session
	ipPool  *infra.IPPool
}

// Deploy deploys environment to tmux target
func (t *TMux) Deploy(ctx context.Context, env infra.Set) error {
	t.session = tmux.NewSession(t.config.EnvName, t.config.LogDir)
	newSession, err := t.session.Init(ctx)
	if err != nil {
		return err
	}
	if newSession {
		if err := env.Deploy(ctx, t); err != nil {
			return err
		}
	}
	if !t.config.TestingMode {
		return t.session.Attach(ctx)
	}
	return nil
}

// DeployBinary starts binary file inside tmux session
func (t *TMux) DeployBinary(ctx context.Context, app infra.Binary) (infra.Deployment, error) {
	var deployment infra.Deployment
	ip, err := t.ipPool.Next()
	if err != nil {
		return deployment, err
	}

	if err := infra.PreprocessApp(ctx, ip, app.AppBase); err != nil {
		return deployment, err
	}
	if err := t.session.StartApp(ctx, app.Name, append([]string{app.Path}, app.Args...)...); err != nil {
		return deployment, err
	}
	deployment.IP = ip
	return deployment, nil
}

// DeployContainer starts container inside tmux session
func (t *TMux) DeployContainer(ctx context.Context, app infra.Container) (infra.Deployment, error) {
	panic("not implemented yet")
}

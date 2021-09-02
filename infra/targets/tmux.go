package targets

import (
	"context"
	"net"

	"github.com/wojciech-sif/localnet/infra"
	"github.com/wojciech-sif/localnet/tmux"
)

// NewTMux creates new tmux target
func NewTMux(config infra.Config, spec *infra.Spec) infra.Target {
	return &TMux{
		config: config,
		spec:   spec,
		ipPool: infra.NewIPPool(config.Network),
	}
}

// TMux is the target deploying apps to tmux session
type TMux struct {
	config  infra.Config
	spec    *infra.Spec
	session *tmux.Session
	ipPool  *infra.IPPool
}

// Stop stops running applications
func (t *TMux) Stop(ctx context.Context) error {
	panic("not implemented")
}

// Deploy deploys environment to tmux target
func (t *TMux) Deploy(ctx context.Context, env infra.Set) error {
	t.session = tmux.NewSession(t.config.EnvName, t.config.LogDir)
	if err := t.session.Init(ctx); err != nil {
		return err
	}
	if err := env.Deploy(ctx, t, t.spec); err != nil {
		return err
	}
	if t.config.TestingMode {
		return nil
	}
	return t.session.Attach(ctx)
}

// DeployBinary starts binary file inside tmux session
func (t *TMux) DeployBinary(ctx context.Context, app infra.Binary) error {
	var ip net.IP
	if app.RequiresIP {
		var err error
		ip, err = t.ipPool.Next()
		if err != nil {
			return err
		}
	}

	if err := infra.PreprocessApp(ctx, ip, t.config.AppDir, app.AppBase); err != nil {
		return err
	}
	if err := t.session.StartApp(ctx, app.Name, append([]string{app.Path}, app.Args...)...); err != nil {
		return err
	}
	return infra.PostprocessApp(ctx, ip, app.AppBase)
}

// DeployContainer starts container inside tmux session
func (t *TMux) DeployContainer(ctx context.Context, app infra.Container) error {
	panic("not implemented yet")
}

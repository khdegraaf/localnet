package targets

import (
	"context"
	"errors"
	"net"
	"sync"

	"github.com/wojciech-sif/localnet/infra"
	"github.com/wojciech-sif/localnet/tmux"
)

// NewTMux creates new tmux target
func NewTMux(config infra.Config) infra.Target {
	return &TMux{
		config: config,
	}
}

// TMux is the target deploying apps to tmux session
type TMux struct {
	config  infra.Config
	session *tmux.Session

	mu        sync.Mutex
	currentIP net.IP
}

// Deploy deploys environment to tmux target
func (t *TMux) Deploy(ctx context.Context, env infra.Env) error {
	t.mu.Lock()
	t.currentIP = t.config.TMuxNetwork
	t.mu.Unlock()

	t.session = tmux.NewSession(t.config.EnvName)
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
	ip, err := t.ip()
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

func (t *TMux) ip() (net.IP, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.currentIP[len(t.currentIP)-1] == 0xfe {
		return nil, errors.New("no more IPs available")
	}
	t.currentIP[len(t.currentIP)-1]++
	ip := make([]byte, len(t.currentIP))
	copy(ip, t.currentIP)
	return ip, nil
}

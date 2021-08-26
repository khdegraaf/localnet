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
func NewTMux(session *tmux.Session, startIP net.IP) *TMux {
	return &TMux{
		session:   session,
		currentIP: startIP,
	}
}

// TMux is the target deploying apps to tmux session
type TMux struct {
	session *tmux.Session

	mu        sync.Mutex
	currentIP net.IP
}

// DeployBinary starts binary file inside tmux session
func (t *TMux) DeployBinary(ctx context.Context, app infra.Binary) (infra.Deployment, error) {
	var deployment infra.Deployment
	ip, err := t.ip()
	if err != nil {
		return deployment, err
	}

	if err := infra.PreprocessApp(ctx, ip.String(), app.AppBase); err != nil {
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

	ip := make([]byte, len(t.currentIP))
	copy(ip, t.currentIP)
	if t.currentIP[len(t.currentIP)-1] == 0xfe {
		return nil, errors.New("no more IPs available")
	}
	t.currentIP[len(t.currentIP)-1]++
	return ip, nil
}

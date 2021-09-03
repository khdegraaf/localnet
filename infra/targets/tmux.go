package targets

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	osexec "os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/ridge/must"
	"github.com/wojciech-sif/localnet/exec"
	"github.com/wojciech-sif/localnet/infra"
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
	config infra.Config
	spec   *infra.Spec
	ipPool *infra.IPPool

	mu sync.Mutex // to protect tmux session
}

// Stop stops running applications
func (t *TMux) Stop(ctx context.Context) error {
	return t.sessionKill(ctx)
}

// Destroy destroys running applications
func (t *TMux) Destroy(ctx context.Context) error {
	return t.Stop(ctx)
}

// Deploy deploys environment to tmux target
func (t *TMux) Deploy(ctx context.Context, env infra.Set) error {
	if err := env.Deploy(ctx, t, t.spec); err != nil {
		return err
	}
	if t.config.TestingMode {
		return nil
	}
	return t.sessionAttach(ctx)
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
	if err := t.sessionAddApp(ctx, app.Name, append([]string{app.Path}, app.Args...)...); err != nil {
		return err
	}
	return infra.PostprocessApp(ctx, ip, app.AppBase)
}

// DeployContainer starts container inside tmux session
func (t *TMux) DeployContainer(ctx context.Context, app infra.Container) error {
	panic("not implemented yet")
}

func (t *TMux) sessionAddApp(ctx context.Context, name string, args ...string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	hasSession, err := t.sessionExists(ctx)
	if err != nil {
		return err
	}
	cmd := []string{
		"bash", "-ce",
		fmt.Sprintf(`exec %s 2>&1 | tee -a "%s/%s.log"`, osexec.Command("", args...).String(), t.config.LogDir, name),
	}
	if hasSession {
		return exec.Run(ctx, exec.TMux(append([]string{"new-window", "-d", "-n", name, "-t", t.config.EnvName + ":"}, cmd...)...))
	}
	return exec.Run(ctx, exec.TMux(append([]string{"new-session", "-d", "-s", t.config.EnvName, "-n", name}, cmd...)...))
}

func (t *TMux) sessionAttach(ctx context.Context) error {
	tty := exec.Tty()
	defer tty.Close()

	return exec.Run(ctx, exec.TMuxTty(tty, "attach-session", "-t", t.config.EnvName))
}

func (t *TMux) sessionKill(ctx context.Context) error {
	// When using just `tmux kill-session` tmux sends SIGHUP to process, but we need SIGTERM.
	// After sending it to all apps, session is terminated automatically.

	t.mu.Lock()
	defer t.mu.Unlock()

	if hasSession, err := t.sessionExists(ctx); err != nil || !hasSession {
		return err
	}
	pids, err := t.sessionPIDs(ctx)
	if err != nil || len(pids) == 0 {
		return err
	}
	return exec.Kill(ctx, pids)
}

func (t *TMux) sessionPIDs(ctx context.Context) ([]int, error) {
	buf := &bytes.Buffer{}
	cmd := exec.TMux("list-windows", "-t", t.config.EnvName, "-F", "#{pane_pid}")
	cmd.Stdout = buf
	if err := exec.Run(ctx, cmd); err != nil {
		return nil, err
	}
	pids := []int{}
	for _, pidStr := range strings.Split(buf.String(), "\n") {
		if pidStr == "" {
			break
		}
		pids = append(pids, int(must.Int64(strconv.ParseInt(pidStr, 10, 32))))
	}
	return pids, nil
}

func (t *TMux) sessionExists(ctx context.Context) (bool, error) {
	err := exec.Run(ctx, exec.TMuxNoOut("has-session", "-t", t.config.EnvName))
	if err != nil && errors.Is(err, ctx.Err()) {
		return false, err
	}
	return err == nil, nil
}

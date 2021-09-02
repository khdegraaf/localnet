package targets

import (
	"context"
	"fmt"
	"net"
	osexec "os/exec"

	"github.com/wojciech-sif/localnet/infra"
)

// NewDirect creates new direct target
func NewDirect(config infra.Config) infra.Target {
	return &Direct{
		config: config,
		ipPool: infra.NewIPPool(config.Network),
	}
}

// Direct is the target deploying apps to os processes
type Direct struct {
	config infra.Config
	ipPool *infra.IPPool
}

// Deploy deploys environment to os processes
func (d *Direct) Deploy(ctx context.Context, env infra.Set) error {
	return env.Deploy(ctx, d)
}

// DeployBinary starts binary file inside os process
func (d *Direct) DeployBinary(ctx context.Context, app infra.Binary) error {
	var ip net.IP
	if app.RequiresIP {
		var err error
		ip, err = d.ipPool.Next()
		if err != nil {
			return err
		}
	}

	if err := infra.PreprocessApp(ctx, ip, d.config.AppDir, app.AppBase); err != nil {
		return err
	}
	cmd := osexec.Command("bash", "-ce", fmt.Sprintf("%s > \"%s/%s.log\" 2>&1", osexec.Command(app.Path, app.Args...).String(), d.config.LogDir, app.Name))
	if err := cmd.Start(); err != nil {
		return err
	}
	return infra.PostprocessApp(ctx, ip, app.AppBase)
}

// DeployContainer starts container
func (d *Direct) DeployContainer(ctx context.Context, app infra.Container) error {
	panic("not implemented yet")
}

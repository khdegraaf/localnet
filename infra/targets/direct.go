package targets

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	osexec "os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"github.com/ridge/must"
	"github.com/wojciech-sif/localnet/infra"
)

// NewDirect creates new direct target
func NewDirect(config infra.Config, spec *infra.Spec) infra.Target {
	return &Direct{
		config: config,
		spec:   spec,
		ipPool: infra.NewIPPool(config.Network),
	}
}

// Direct is the target deploying apps to os processes
type Direct struct {
	config infra.Config
	spec   *infra.Spec
	ipPool *infra.IPPool
}

// Deploy deploys environment to os processes
func (d *Direct) Deploy(ctx context.Context, env infra.Set) error {
	return env.Deploy(ctx, d, d.spec)
}

// Stop stops running applications
func (d *Direct) Stop(ctx context.Context) error {
	if d.spec.PGID == 0 {
		return nil
	}
	procs, err := ioutil.ReadDir("/proc")
	if err != nil {
		return err
	}
	reg := regexp.MustCompile("^[0-9]+$")
	for _, procH := range procs {
		if !procH.IsDir() || !reg.MatchString(procH.Name()) {
			continue
		}
		statPath := "/proc/" + procH.Name() + "/stat"
		statRaw, err := ioutil.ReadFile(statPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return err
		}
		properties := strings.Split(string(statRaw), " ")
		pgID, err := strconv.ParseInt(properties[4], 10, 32)
		if err != nil {
			return err
		}
		if pgID != int64(d.spec.PGID) {
			continue
		}
		pID := int(must.Int64(strconv.ParseInt(procH.Name(), 10, 32)))
		if pID == os.Getpid() {
			continue
		}

		proc, err := os.FindProcess(pID)
		if err != nil {
			continue
		}

		// FIXME (wojciech): send sigkill after timeout
		// FIXME (wojciech): parallelize it
		// FIXME (wojciech): wait until process is done
		if err := proc.Signal(syscall.SIGTERM); err != nil {
			return err
		}
	}
	return nil
}

// Destroy destroys running applications
func (d *Direct) Destroy(ctx context.Context) error {
	return d.Stop(ctx)
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
	cmd := osexec.Command("bash", "-ce", fmt.Sprintf("%s >> \"%s/%s.log\" 2>&1", osexec.Command(app.Path, app.Args...).String(), d.config.LogDir, app.Name))
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
		Pgid:    d.spec.PGID,
	}

	if err := cmd.Start(); err != nil {
		return err
	}
	return infra.PostprocessApp(ctx, ip, app.AppBase)
}

// DeployContainer starts container
func (d *Direct) DeployContainer(ctx context.Context, app infra.Container) error {
	panic("not implemented yet")
}

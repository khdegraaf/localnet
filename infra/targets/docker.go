package targets

import (
	"bytes"
	"context"
	"fmt"
	"net"
	osexec "os/exec"
	"strings"
	"text/template"

	"github.com/wojciech-sif/localnet/exec"
	"github.com/wojciech-sif/localnet/infra"
)

const dockerTplContent = `FROM fedora:latest
{{ range .Copy }}
COPY {{ . }} {{ . }}
{{ end }}
ENTRYPOINT ["{{ .Path }}"]
`

var dockerTpl = template.Must(template.New("").Parse(dockerTplContent))

// NewDocker creates new docker target
func NewDocker(config infra.Config) infra.Target {
	return &Docker{
		config: config,
	}
}

// Docker is the target deploying apps to docker
type Docker struct {
	config infra.Config
}

// Deploy deploys environment to docker target
func (d *Docker) Deploy(ctx context.Context, env infra.Set) error {
	return env.Deploy(ctx, d)
}

// DeployBinary builds container image out of binary file and starts it in docker
func (d *Docker) DeployBinary(ctx context.Context, app infra.Binary) (infra.Deployment, error) {
	var deployment infra.Deployment

	if err := infra.PreprocessApp(ctx, net.IPv4zero, app.AppBase); err != nil {
		return deployment, err
	}

	buf := &bytes.Buffer{}
	if err := dockerTpl.Execute(buf, app); err != nil {
		return deployment, err
	}

	image := app.Name + ":latest"
	buildCmd := exec.Docker("build", "--tag", image, "-f-", "/")
	buildCmd.Stdin = buf

	name := d.config.EnvName + "-" + app.Name
	ipBuf := &bytes.Buffer{}
	ipCmd := exec.Docker("inspect", "-f", "{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}", name)
	ipCmd.Stdout = ipBuf
	err := exec.Run(ctx,
		buildCmd,
		exec.Docker(append([]string{"run", "--name", name, "-d", image}, app.Args...)...),
		ipCmd,
	)
	if err != nil {
		return deployment, err
	}

	err = osexec.Command("bash", "-ce",
		fmt.Sprintf("%s > \"%s/%s.log\" 2>&1", exec.Docker("logs", "-f", name).String(),
			d.config.LogDir, app.Name)).Start()
	if err != nil {
		return deployment, err
	}

	deployment.IP = net.ParseIP(strings.TrimSuffix(ipBuf.String(), "\n"))
	return deployment, nil
}

// DeployContainer starts container in docker
func (d *Docker) DeployContainer(ctx context.Context, app infra.Container) (infra.Deployment, error) {
	panic("not implemented yet")
}

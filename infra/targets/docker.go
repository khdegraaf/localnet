package targets

import (
	"bytes"
	"context"
	"net"
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
func NewDocker() *Docker {
	return &Docker{}
}

// Docker is the target deploying apps to docker
type Docker struct {
}

// DeployBinary builds container image out of binary file and starts it in docker
func (d *Docker) DeployBinary(ctx context.Context, app infra.Binary) (infra.Deployment, error) {
	var deployment infra.Deployment

	if err := infra.PreprocessApp(ctx, "0.0.0.0", app.AppBase); err != nil {
		return deployment, err
	}

	buf := &bytes.Buffer{}
	if err := dockerTpl.Execute(buf, app); err != nil {
		return deployment, err
	}

	image := app.Name + ":latest"
	buildCmd := exec.Docker("build", "--tag", image, "-f-", "/")
	buildCmd.Stdin = buf

	ipBuf := &bytes.Buffer{}
	ipCmd := exec.Docker("inspect", "-f", "{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}", app.Name)
	ipCmd.Stdout = ipBuf
	err := exec.Run(ctx,
		buildCmd,
		exec.Docker(append([]string{"run", "--name", app.Name, "-d", image}, app.Args...)...),
		exec.Docker("inspect", "-f", "{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}", app.Name),
		ipCmd,
	)
	if err != nil {
		return deployment, err
	}
	deployment.IP = net.ParseIP(strings.TrimSuffix(ipBuf.String(), "\n")).To4()
	return deployment, nil
}

// DeployContainer starts container in docker
func (d *Docker) DeployContainer(ctx context.Context, app infra.Container) (infra.Deployment, error) {
	panic("not implemented yet")
}

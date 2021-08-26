package targets

import (
	"context"

	"github.com/wojciech-sif/localnet/infra"
)

// Docker is the target deploying apps to docker
type Docker struct {
}

// DeployBinary builds container image out of binary file and starts it in docker
func (d *Docker) DeployBinary(ctx context.Context, app infra.Binary) (infra.Deployment, error) {
	panic("not implemented yet")
}

// DeployContainer starts container in docker
func (d *Docker) DeployContainer(ctx context.Context, app infra.Container) (infra.Deployment, error) {
	panic("not implemented yet")
}

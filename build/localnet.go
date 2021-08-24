package build

import (
	"context"
	osexec "os/exec"

	"github.com/wojciech-malota-wojcik/build"
	"github.com/wojciech-sif/localnet/exec"
)

func buildApp(ctx context.Context) error {
	return goBuildPkg(ctx, "cmd", "bin/localnet", false)
}

func runApp(ctx context.Context, deps build.DepsFunc) error {
	deps(buildApp)
	return exec.Run(ctx, osexec.Command("./bin/localnet"))
}

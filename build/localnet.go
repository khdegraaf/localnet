package build

import (
	"context"
	"os/exec"

	"github.com/wojciech-malota-wojcik/build"
)

func buildApp(ctx context.Context) error {
	return goBuildPkg(ctx, "cmd", "bin/localnet", false)
}

func runApp(ctx context.Context, deps build.DepsFunc) error {
	deps(buildApp)
	return runCmd(ctx, exec.Command("./bin/localnet"))
}

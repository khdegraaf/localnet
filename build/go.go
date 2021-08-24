package build

import (
	"context"
	"os"
	osexec "os/exec"

	"github.com/wojciech-malota-wojcik/build"
	"github.com/wojciech-sif/localnet/exec"
)

func goBuildPkg(ctx context.Context, pkg, out string, cgo bool) error {
	cmd := osexec.Command("go", "build", "-o", out, "./"+pkg)
	if !cgo {
		cmd.Env = append([]string{"CGO_ENABLED=0"}, os.Environ()...)
	}
	return exec.Run(ctx, cmd)
}

func goModTidy(ctx context.Context) error {
	return exec.Run(ctx, osexec.Command("go", "mod", "tidy"))
}

func goLint(ctx context.Context, deps build.DepsFunc) error {
	if err := exec.Run(ctx, osexec.Command("golangci-lint", "run", "--config", "build/.golangci.yaml")); err != nil {
		return err
	}
	deps(goModTidy, gitStatusClean)
	return nil
}

func goImports(ctx context.Context) error {
	return exec.Run(ctx, osexec.Command("goimports", "-w", "."))
}

func goTest(ctx context.Context) error {
	return exec.Run(ctx, osexec.Command("go", "test", "-count=1", "-race", "./..."))
}

package build

import (
	"context"
	"os"
	"os/exec"
	"syscall"

	"github.com/ridge/parallel"
	"github.com/wojciech-malota-wojcik/build"
)

func runCmd(ctx context.Context, cmd *exec.Cmd) error {
	if cmd.Stdout == nil {
		cmd.Stdout = os.Stdout
	}
	if cmd.Stderr == nil {
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	g := parallel.NewGroup(ctx)
	g.Spawn("cmd", parallel.Exit, func(ctx context.Context) error {
		err := cmd.Wait()
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return err
	})
	g.Spawn("ctx", parallel.Exit, func(ctx context.Context) error {
		<-ctx.Done()
		_ = cmd.Process.Signal(syscall.SIGTERM)
		return ctx.Err()
	})
	return g.Wait()
}

func goBuildPkg(ctx context.Context, pkg, out string, cgo bool) error {
	cmd := exec.Command("go", "build", "-o", out, "./"+pkg)
	if !cgo {
		cmd.Env = append([]string{"CGO_ENABLED=0"}, os.Environ()...)
	}
	return runCmd(ctx, cmd)
}

func goModTidy(ctx context.Context) error {
	return runCmd(ctx, exec.Command("go", "mod", "tidy"))
}

func goLint(ctx context.Context, deps build.DepsFunc) error {
	if err := runCmd(ctx, exec.Command("golangci-lint", "run", "--config", "build/.golangci.yaml")); err != nil {
		return err
	}
	deps(goModTidy, gitStatusClean)
	return nil
}

func goImports(ctx context.Context) error {
	return runCmd(ctx, exec.Command("goimports", "-w", "."))
}

func goTest(ctx context.Context) error {
	return runCmd(ctx, exec.Command("go", "test", "-count=1", "-race", "./..."))
}

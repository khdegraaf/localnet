package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/ridge/must"
	"github.com/wojciech-malota-wojcik/build"
	"github.com/wojciech-malota-wojcik/ioc"
	me "github.com/wojciech-sif/localnet/build"
	"github.com/wojciech-sif/localnet/lib/run"
)

func main() {
	run.Tool("build", nil, func(ctx context.Context, c *ioc.Container) error {
		exec := build.NewIoCExecutor(me.Commands, c)
		if build.Autocomplete(exec) {
			return nil
		}

		changeWorkingDir()
		return build.Do(ctx, "CMS", exec)
	})
}

func changeWorkingDir() {
	must.OK(os.Chdir(filepath.Dir(filepath.Dir(must.String(filepath.EvalSymlinks(must.String(os.Executable())))))))
}

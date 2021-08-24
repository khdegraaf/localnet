package main

import (
	"context"

	"github.com/wojciech-malota-wojcik/build"
	"github.com/wojciech-malota-wojcik/ioc"
	me "github.com/wojciech-sif/localnet/build"
	"github.com/wojciech-sif/localnet/lib/run"
)

func main() {
	run.Tool("build", func(appRunner run.AppRunner, c *ioc.Container) {
		exec := build.NewIoCExecutor(me.Commands, c)
		if build.Autocomplete(exec) {
			return
		}
		appRunner(func(ctx context.Context) error {
			return build.Do(ctx, "CMS", exec)
		})
	})
}

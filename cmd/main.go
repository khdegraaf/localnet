package main

import (
	"context"

	"github.com/wojciech-malota-wojcik/ioc"
	"github.com/wojciech-sif/localnet/lib/run"
	"github.com/wojciech-sif/localnet/tmux"
)

func main() {
	run.Tool("localnet", func(appRunner run.AppRunner, c *ioc.Container) {
		appRunner(func(ctx context.Context) error {
			// FIXME (wojciech): pass session name from CLI
			session := tmux.NewSession("localnet")
			session.AddWindow("bash1", "bash")
			session.AddWindow("bash2", "bash")
			return session.Run(ctx)
		})
	})
}

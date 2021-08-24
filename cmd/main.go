package main

import (
	"context"

	"github.com/ridge/must"
	"github.com/wojciech-malota-wojcik/ioc"
	"github.com/wojciech-sif/localnet/exec"
	"github.com/wojciech-sif/localnet/lib/run"
)

const sessionName = "localnet"

func main() {
	run.Tool("localnet", func(appRunner run.AppRunner, c *ioc.Container) {
		appRunner(func(ctx context.Context) error {
			if exec.TmateNoOut(ctx, "has-session", "-t", sessionName) != nil {
				must.OK(exec.Tmate(ctx, "new-session", "-d", "-s", sessionName, "-n", "help", "bash"))
			}
			must.OK(exec.TmateTty(ctx, "attach-session", "-t", sessionName))
			return nil
		})
	})
}

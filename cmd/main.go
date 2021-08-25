package main

import (
	"context"

	"github.com/wojciech-malota-wojcik/ioc"
	"github.com/wojciech-sif/localnet/infra"
	"github.com/wojciech-sif/localnet/infra/targets"
	"github.com/wojciech-sif/localnet/lib/run"
	"github.com/wojciech-sif/localnet/tmux"
)

type app interface {
	Deploy(ctx context.Context, target infra.Target) error
}

type apps []app

func (apps apps) Deploy(ctx context.Context, t infra.Target) error {
	for _, app := range apps {
		if err := app.Deploy(ctx, t); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	run.Tool("localnet", func(appRunner run.AppRunner, c *ioc.Container) {
		appRunner(func(ctx context.Context) error {
			// FIXME (wojciech): pass session name from CLI
			session := tmux.NewSession("localnet")
			newSession, err := session.Init(ctx)
			if err != nil {
				return err
			}
			if newSession {
				target := targets.NewTMux(session)
				apps := apps{
					infra.Binary{
						Name: "bash1",
						Path: "bash",
					},
					infra.Binary{
						Name: "bash2",
						Path: "bash",
					},
				}
				if err := apps.Deploy(ctx, target); err != nil {
					return err
				}
			}
			return session.Attach(ctx)
		})
	})
}

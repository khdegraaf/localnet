package main

import (
	"context"

	"github.com/wojciech-malota-wojcik/ioc"
	"github.com/wojciech-sif/localnet/infra"
	"github.com/wojciech-sif/localnet/infra/apps"
	"github.com/wojciech-sif/localnet/infra/targets"
	"github.com/wojciech-sif/localnet/lib/run"
	"github.com/wojciech-sif/localnet/tmux"
)

type app interface {
	Deploy(ctx context.Context, target infra.Target) error
}

type env []app

func (e env) Deploy(ctx context.Context, t infra.Target) error {
	for _, app := range e {
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
				env := env{
					apps.NewSifchain("sifchain-a", "127.0.0.1"),
					apps.NewSifchain("sifchain-b", "127.0.0.2"),
					apps.NewHermes("hermes", "127.0.0.3", "sifchain-a", "127.0.0.1", "sifchain-b", "127.0.0.2"),
				}
				if err := env.Deploy(ctx, target); err != nil {
					return err
				}
			}
			return session.Attach(ctx)
		})
	})
}

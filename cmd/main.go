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

func env() infra.Env {
	sifchainA := apps.NewSifchain("sifchain-a", "127.0.0.1")
	sifchainB := apps.NewSifchain("sifchain-b", "127.0.0.2")
	return infra.Env{
		sifchainA,
		sifchainB,
		apps.NewHermes("hermes", "127.0.0.3", sifchainA, sifchainB),
	}
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
				if err := env().Deploy(ctx, target); err != nil {
					return err
				}
			}
			return session.Attach(ctx)
		})
	})
}

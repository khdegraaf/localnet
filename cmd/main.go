package main

import (
	"context"
	"os"

	"github.com/ridge/must"
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

// FIXME (wojciech): read it from CLI
const envName = "default"

func main() {
	run.Tool("localnet", func(appRunner run.AppRunner, c *ioc.Container) {
		appRunner(func(ctx context.Context) error {
			// FIXME (wojciech): read config from CLI
			config := infra.Config{
				EnvName: envName,
				HomeDir: must.String(os.UserHomeDir()) + "/.localnet/" + envName,
			}

			session := tmux.NewSession("localnet-" + config.EnvName)
			newSession, err := session.Init(ctx)
			if err != nil {
				return err
			}
			if newSession {
				target := targets.NewTMux(session)
				if err := env().Deploy(ctx, config, target); err != nil {
					return err
				}
			}
			return session.Attach(ctx)
		})
	})
}

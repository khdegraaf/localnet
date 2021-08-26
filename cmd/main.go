package main

import (
	"context"

	"github.com/wojciech-sif/localnet"
	"github.com/wojciech-sif/localnet/infra"
	"github.com/wojciech-sif/localnet/infra/apps"
	"github.com/wojciech-sif/localnet/lib/run"
)

func env() infra.Env {
	sifchainA := apps.NewSifchain("sifchain-a")
	sifchainB := apps.NewSifchain("sifchain-b")
	return infra.Env{
		sifchainA,
		sifchainB,
		apps.NewHermes("hermes", sifchainA, sifchainB),
	}
}

func main() {
	run.Tool("localnet", localnet.IoC, func(ctx context.Context, target infra.Target) error {
		return target.Deploy(ctx, env())
	})
}

package localnet

import (
	"github.com/wojciech-sif/localnet/infra"
	"github.com/wojciech-sif/localnet/infra/apps"
)

// DevEnv is the environment for developer
func DevEnv() infra.Env {
	sifchainA := apps.NewSifchain("sifchain-a")
	sifchainB := apps.NewSifchain("sifchain-b")
	return infra.Env{
		sifchainA,
		sifchainB,
		apps.NewHermes("hermes", sifchainA, sifchainB),
	}
}

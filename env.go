package localnet

import (
	"github.com/wojciech-sif/localnet/infra"
	"github.com/wojciech-sif/localnet/infra/apps"
)

// DevEnv is the environment for developer
func DevEnv(af *apps.Factory) infra.Env {
	sifchainA := af.Sifchain("sifchain-a")
	sifchainB := af.Sifchain("sifchain-b")
	return infra.Env{
		sifchainA,
		sifchainB,
		af.Hermes("hermes", sifchainA, sifchainB),
	}
}

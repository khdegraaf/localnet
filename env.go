package localnet

import (
	"github.com/wojciech-sif/localnet/infra"
	"github.com/wojciech-sif/localnet/infra/apps"
)

// DevSet is the environment for developer
func DevSet(af *apps.Factory) infra.Set {
	sifchainA := af.Sifchain("sifchain-a")
	sifchainB := af.Sifchain("sifchain-b")
	return infra.Set{
		sifchainA,
		sifchainB,
		af.Hermes("hermes", sifchainA, sifchainB),
	}
}

// SingleChainSet is the environment with one sifchain
func SingleChainSet(af *apps.Factory) infra.Set {
	return infra.Set{
		af.Sifchain("sifchain"),
	}
}

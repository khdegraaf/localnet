package localnet

import (
	"github.com/wojciech-sif/localnet/infra"
	"github.com/wojciech-sif/localnet/infra/apps"
	"github.com/wojciech-sif/localnet/tests"
)

// DevSet is the environment for developer
func DevSet(af *apps.Factory) infra.Set {
	return infra.Set{
		af.Sifchain("sifchain"),
	}
}

// FullSet is the environment with all apps
func FullSet(af *apps.Factory) infra.Set {
	sifchainA := af.Sifchain("sifchain-a")
	sifchainB := af.Sifchain("sifchain-b")
	return infra.Set{
		sifchainA,
		sifchainB,
		af.Hermes("hermes", sifchainA, sifchainB),
	}
}

// TestsSet returns environment used for testing
func TestsSet(af *apps.Factory) infra.Set {
	env, _ := tests.Tests(af)
	return env
}

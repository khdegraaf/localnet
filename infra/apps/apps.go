package apps

import (
	"github.com/wojciech-sif/localnet/infra"
	"github.com/wojciech-sif/localnet/infra/apps/hermes"
	"github.com/wojciech-sif/localnet/infra/apps/sifchain"
)

// NewFactory creates new app factory
func NewFactory(config infra.Config, spec *infra.Spec) *Factory {
	return &Factory{
		config: config,
		spec:   spec,
	}
}

// Factory produces apps from config
type Factory struct {
	config infra.Config
	spec   *infra.Spec
}

// Sifchain creates new sifchain
func (f *Factory) Sifchain(name string) *Sifchain {
	return NewSifchain(f.config.WrapperDir, sifchain.NewExecutor(name, f.config.BinDir+"/sifnoded", f.config.AppDir+"/"+name,
		"master"), f.spec)
}

// Hermes creates new hermes
func (f *Factory) Hermes(name string, chainA, chainB hermes.Peer) *Hermes {
	return NewHermes(f.config, "hermes", f.spec, chainA, chainB)
}

package apps

import (
	"github.com/wojciech-sif/localnet/infra"
	"github.com/wojciech-sif/localnet/infra/apps/hermes"
	"github.com/wojciech-sif/localnet/infra/apps/sifchain"
)

// NewFactory creates new app factory
func NewFactory(config infra.Config) *Factory {
	return &Factory{
		config: config,
	}
}

// Factory produces apps from config
type Factory struct {
	config infra.Config
}

// Sifchain creates new sifchain
func (f *Factory) Sifchain(name string) *Sifchain {
	return NewSifchain(f.config.WrapperDir, sifchain.NewExecutor(name, f.config.BinDir+"/sifnoded", f.config.AppDir+"/"+name,
		"master"))
}

// Hermes creates new hermes
func (f *Factory) Hermes(name string, chainA, chainB hermes.Peer) *Hermes {
	return NewHermes(f.config, "hermes", chainA, chainB)
}

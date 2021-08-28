package tests

import (
	"github.com/wojciech-sif/localnet/infra"
	"github.com/wojciech-sif/localnet/infra/apps"
	"github.com/wojciech-sif/localnet/infra/testing"
	"github.com/wojciech-sif/localnet/tests/transfers"
)

// Tests returns testing environment and tests
func Tests(appF *apps.Factory) (infra.Set, []*testing.T) {
	chain := appF.Sifchain("sifchain")
	return infra.Set{
			chain,
		},
		[]*testing.T{
			testing.New(transfers.VerifyInitialBalance(chain)),
			testing.New(transfers.TransferRowan(chain)),
		}
}

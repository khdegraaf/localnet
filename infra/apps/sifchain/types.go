package sifchain

import (
	"context"
	"math/big"
	"sync"

	"github.com/wojciech-sif/localnet/lib/rnd"
)

// Wallet stores information related to wallet
type Wallet struct {
	// Name is the name of the key stored in keystore
	Name string

	// Address is the address of the wallet
	Address string
}

// Balance stores balance of denom
type Balance struct {
	// Amount is stored amount
	Amount *big.Int

	// Denom is a token symbol
	Denom string
}

// NewGenesis returns new genesis configurator
func NewGenesis(executor *Executor) *Genesis {
	return &Genesis{
		executor: executor,
		wallets:  map[Wallet][]Balance{},
	}
}

// Genesis represents configuration of genesis block
type Genesis struct {
	executor *Executor

	mu      sync.Mutex
	wallets map[Wallet][]Balance
}

// AddWallet adds wallet with balances to the genesis
func (g *Genesis) AddWallet(ctx context.Context, balances ...Balance) (Wallet, error) {
	name := rnd.GetRandomName()
	addr, _, err := g.executor.AddKey(ctx, name)
	if err != nil {
		return Wallet{}, err
	}
	wallet := Wallet{Name: name, Address: addr}

	g.mu.Lock()
	defer g.mu.Unlock()

	g.wallets[wallet] = balances

	return wallet, nil
}

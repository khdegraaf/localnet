package sifchain

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net"
	osexec "os/exec"
	"strings"

	"github.com/wojciech-sif/localnet/exec"
)

// NewExecutor returns new executor
func NewExecutor(name, binPath, homeDir, keyName string) *Executor {
	return &Executor{
		name:    name,
		binPath: binPath,
		homeDir: homeDir,
		keyName: keyName,
	}
}

// Executor exposes methods for executing sifnoded binary
type Executor struct {
	name    string
	binPath string
	homeDir string
	keyName string
}

// Name returns name of the chain
func (e *Executor) Name() string {
	return e.name
}

// Bin returns path to sifnode binary
func (e *Executor) Bin() string {
	return e.binPath
}

// Home returns path to home dir
func (e *Executor) Home() string {
	return e.homeDir
}

// AddKey adds key to the client
func (e *Executor) AddKey(ctx context.Context, name string) (addr, validatorAddr string, err error) {
	keyData := &bytes.Buffer{}
	addrBuf := &bytes.Buffer{}
	validatorAddrBuf := &bytes.Buffer{}
	err = exec.Run(ctx,
		e.sifnodedOut(keyData, "keys", "add", name, "--output", "json", "--keyring-backend", "test"),
		e.sifnodedOut(addrBuf, "keys", "show", name, "-a", "--keyring-backend", "test"),
		e.sifnodedOut(validatorAddrBuf, "keys", "show", name, "-a", "--bech", "val", "--keyring-backend", "test"),
	)
	if err != nil {
		return "", "", err
	}
	return strings.TrimSuffix(addrBuf.String(), "\n"), strings.TrimSuffix(validatorAddrBuf.String(), "\n"), ioutil.WriteFile(e.homeDir+"/"+name+".json", keyData.Bytes(), 0o600)
}

// PrepareNode prepares node to start
func (e *Executor) PrepareNode(ctx context.Context, genesis *Genesis) error {
	addr, valAddr, err := e.AddKey(ctx, e.keyName)
	if err != nil {
		return err
	}

	cmds := []*osexec.Cmd{
		e.sifnoded("init", e.name, "--chain-id", e.name, "-o"),
		e.sifnoded("add-genesis-account", addr, "500000000000000000000000rowan,990000000000000000000000000stake", "--keyring-backend", "test"),
		e.sifnoded("add-genesis-validators", valAddr, "--keyring-backend", "test"),
	}
	for wallet, balances := range genesis.wallets {
		if len(balances) == 0 {
			continue
		}
		balancesStr := ""
		for _, balance := range balances {
			if balancesStr != "" {
				balancesStr += ","
			}
			balancesStr += balance.Amount.String() + balance.Denom
		}
		cmds = append(cmds, e.sifnoded("add-genesis-account", wallet.Address, balancesStr, "--keyring-backend", "test"))
	}
	cmds = append(cmds,
		e.sifnoded("gentx", e.keyName, "1000000000000000000000000stake", "--chain-id", e.name, "--keyring-backend", "test"),
		e.sifnoded("collect-gentxs"),
	)
	return exec.Run(ctx, cmds...)
}

// QBankBalances queries for bank balances owned by address
func (e *Executor) QBankBalances(ctx context.Context, address string, ip net.IP) ([]byte, error) {
	balances := &bytes.Buffer{}
	if err := exec.Run(ctx, e.sifnodedOut(balances, "q", "bank", "balances", address, "--chain-id", e.name, "--node", fmt.Sprintf("tcp://%s:26657", ip), "--output", "json")); err != nil {
		return nil, err
	}
	return balances.Bytes(), nil
}

// TxBankSend sends tokens from one address to another
func (e *Executor) TxBankSend(ctx context.Context, sender, address string, balance Balance, ip net.IP) ([]byte, error) {
	tx := &bytes.Buffer{}
	if err := exec.Run(ctx, e.sifnodedOut(tx, "tx", "bank", "send", sender, address, balance.Amount.String()+balance.Denom, "--yes", "--chain-id", e.name, "--node", fmt.Sprintf("tcp://%s:26657", ip), "--keyring-backend", "test", "--output", "json")); err != nil {
		return nil, err
	}
	return tx.Bytes(), nil
}

func (e *Executor) sifnoded(args ...string) *osexec.Cmd {
	return osexec.Command(e.binPath, append([]string{"--home", e.homeDir}, args...)...)
}

func (e *Executor) sifnodedOut(buf *bytes.Buffer, args ...string) *osexec.Cmd {
	cmd := e.sifnoded(args...)
	cmd.Stdout = buf
	return cmd
}

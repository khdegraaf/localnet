package apps

import (
	"context"
	"errors"
	"io/ioutil"
	"net"
	"os"
	"sync"

	"github.com/ridge/must"
	"github.com/wojciech-sif/localnet/infra"
	"github.com/wojciech-sif/localnet/infra/apps/sifchain"
)

// NewSifchain creates new sifchain app
func NewSifchain(homeDir string, executor *sifchain.Executor) *Sifchain {
	if err := os.RemoveAll(executor.Home()); err != nil && !errors.Is(err, os.ErrNotExist) {
		panic(err)
	}
	must.OK(os.MkdirAll(executor.Home(), 0o700))

	return &Sifchain{
		homeDir:  homeDir,
		executor: executor,
		genesis:  sifchain.NewGenesis(executor),
	}
}

// Sifchain represents sifchain
type Sifchain struct {
	homeDir  string
	executor *sifchain.Executor
	genesis  *sifchain.Genesis

	mu sync.RWMutex
	ip net.IP
}

// ID returns chain ID
func (s *Sifchain) ID() string {
	return s.executor.Name()
}

// IP returns IP chain listens on
func (s *Sifchain) IP() net.IP {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.ip
}

// Genesis returns configurator of genesis block
func (s *Sifchain) Genesis() *sifchain.Genesis {
	return s.genesis
}

// Client creates new client for sifchain blockchain
func (s *Sifchain) Client() *sifchain.Client {
	return sifchain.NewClient(s.executor, s.ip)
}

// Deploy deploys sifchain app to the target
func (s *Sifchain) Deploy(ctx context.Context, target infra.AppTarget) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	deployment, err := target.DeployBinary(ctx, infra.Binary{
		Path: s.executor.Bin(),
		AppBase: infra.AppBase{
			Name: s.executor.Name(),
			Args: []string{
				"start",
				"--home", s.executor.Home(),
				"--rpc.laddr", "tcp://{{ .IP }}:26657",
				"--p2p.laddr", "tcp://{{ .IP }}:26656",
				"--grpc.address", "{{ .IP }}:9090",
				"--rpc.pprof_laddr", "{{ .IP }}:6060",
			},
			Copy: []string{
				s.executor.Bin(),
				s.executor.Home(),
			},
			PreFunc: func(ctx context.Context) error {
				return s.executor.PrepareNode(ctx, s.genesis)
			},
		},
	})
	if err != nil {
		return err
	}
	s.ip = deployment.IP
	return s.saveClientWrapper(s.homeDir)
}

func (s *Sifchain) saveClientWrapper(home string) error {
	client := `#!/bin/sh
OPTS=""
if [ "$1" == "tx" ] || [ "$1" == "q" ]; then
	OPTS="$OPTS --chain-id ""` + s.executor.Name() + `"" --node ""tcp://` + s.ip.String() + `:26657"""
fi
if [ "$1" == "tx" ] || [ "$1" == "keys" ]; then
	OPTS="$OPTS --keyring-backend ""test"""
fi

exec ` + s.executor.Bin() + ` --home "` + s.executor.Home() + `" "$@" $OPTS
`

	if err := os.MkdirAll(home+"/bin", 0o700); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}
	return ioutil.WriteFile(home+"/bin/"+s.executor.Name(), []byte(client), 0o700)
}

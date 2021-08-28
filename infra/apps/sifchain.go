package apps

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/ridge/must"
	"github.com/wojciech-sif/localnet/infra"
	"github.com/wojciech-sif/localnet/infra/apps/sifchain"
	"github.com/wojciech-sif/localnet/lib/retry"
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

// HealthCheck checks if sifchain is empty
func (s *Sifchain) HealthCheck(ctx context.Context) error {
	if s.ip == nil {
		return retry.Retryable(fmt.Errorf("sifchain hasn't started yet"))
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	req := must.HTTPRequest(http.NewRequestWithContext(ctx, http.MethodGet, "http://"+s.ip.String()+":26657/status", nil))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return retry.Retryable(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return retry.Retryable(err)
	}

	if resp.StatusCode != http.StatusOK {
		return retry.Retryable(fmt.Errorf("health check failed, status code: %d, response: %s", resp.StatusCode, body))
	}

	data := struct {
		Result struct {
			SyncInfo struct {
				LatestBlockHash string `json:"latest_block_hash"` // nolint: tagliatelle
			} `json:"sync_info"` // nolint: tagliatelle
		} `json:"result"`
	}{}

	if err := json.Unmarshal(body, &data); err != nil {
		return retry.Retryable(err)
	}

	if data.Result.SyncInfo.LatestBlockHash == "" {
		return retry.Retryable(errors.New("genesis block hasn't been mined yet"))
	}

	return nil
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

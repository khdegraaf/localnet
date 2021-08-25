package apps

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	osexec "os/exec"
	"time"

	"github.com/ridge/must"
	"github.com/wojciech-sif/localnet/exec"
	"github.com/wojciech-sif/localnet/infra"
)

// SifchainPeer is the interface required by hermes to be able to connect to sifchain
type SifchainPeer interface {
	// ID returns chain id
	ID() string

	// IP returns ip used for connection
	IP() string
}

// NewHermes creates new hermes app
func NewHermes(name, ip string, chainA, chainB SifchainPeer) *Hermes {
	return &Hermes{
		name:   name,
		ip:     ip,
		chainA: chainA,
		chainB: chainB,
	}
}

// Hermes represents hermes relayer
type Hermes struct {
	name   string
	ip     string
	chainA SifchainPeer
	chainB SifchainPeer
}

// Deploy deploys sifchain app to the target
func (h *Hermes) Deploy(ctx context.Context, config infra.Config, target infra.Target) error {
	// FIXME (wojciech): implement healthchecks instead of this hack
	time.Sleep(10 * time.Second)

	hermesHome := config.HomeDir + "/" + h.name
	configFile := hermesHome + "/config.toml"
	hermes := func(args ...string) *osexec.Cmd {
		return osexec.Command("hermes", append([]string{"--config", configFile}, args...)...)
	}

	if err := os.RemoveAll(hermesHome); err != nil && !errors.Is(err, os.ErrNotExist) {
		panic(err)
	}
	must.OK(os.MkdirAll(hermesHome, 0o700))

	cfg := `[global]
strategy = 'packets'
filter = false
log_level = 'info'
clear_packets_interval = 100

[telemetry]
enabled = true
host = '` + h.ip + `'
port = 3001

[[chains]]
id = '` + h.chainA.ID() + `'
rpc_addr = 'http://` + h.chainA.IP() + `:26657'
grpc_addr = 'http://` + h.chainA.IP() + `:9090'
websocket_addr = 'ws://` + h.chainA.IP() + `:26657/websocket'
rpc_timeout = '10s'
account_prefix = 'sif'
key_name = '` + h.chainA.ID() + `'
store_prefix = 'ibc'
max_gas = 3000000
gas_price = { price = 0.001, denom = 'stake' }
gas_adjustment = 0.1
max_msg_num = 30
max_tx_size = 2097152
clock_drift = '5s'
trusting_period = '14days'
trust_threshold = { numerator = '1', denominator = '3' }

[[chains]]
id = '` + h.chainB.ID() + `'
rpc_addr = 'http://` + h.chainB.IP() + `:26657'
grpc_addr = 'http://` + h.chainB.IP() + `:9090'
websocket_addr = 'ws://` + h.chainB.IP() + `:26657/websocket'
rpc_timeout = '10s'
account_prefix = 'sif'
key_name = '` + h.chainB.ID() + `'
store_prefix = 'ibc'
max_gas = 3000000
gas_price = { price = 0.001, denom = 'stake' }
gas_adjustment = 0.1
max_msg_num = 30
max_tx_size = 2097152
clock_drift = '5s'
trusting_period = '14days'
trust_threshold = { numerator = '1', denominator = '3' }
`
	must.OK(ioutil.WriteFile(configFile, []byte(cfg), 0o600))

	err := exec.Run(ctx,
		hermes("keys", "add", h.chainA.ID(), "--file", config.HomeDir+"/"+h.chainA.ID()+".json"),
		hermes("keys", "add", h.chainB.ID(), "--file", config.HomeDir+"/"+h.chainB.ID()+".json"),
		hermes("create", "channel", h.chainA.ID(), h.chainB.ID(), "--port-a", "transfer", "--port-b", "transfer"),
	)
	if err != nil {
		return err
	}
	return target.DeployBinary(ctx, infra.Binary{
		Name: h.name,
		Path: "hermes",
		Args: []string{
			"--config", configFile,
			"start",
		},
	})
}

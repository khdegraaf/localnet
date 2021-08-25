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

// NewHermes creates new hermes app
func NewHermes(name, ip, chainID1, ip1, chainID2, ip2 string) *Hermes {
	return &Hermes{
		name:     name,
		ip:       ip,
		chainID1: chainID1,
		chainID2: chainID2,
		ip1:      ip1,
		ip2:      ip2,
	}
}

// Hermes represents hermes relayer
type Hermes struct {
	name     string
	ip       string
	chainID1 string
	chainID2 string
	ip1      string
	ip2      string
}

// Deploy deploys sifchain app to the target
func (h *Hermes) Deploy(ctx context.Context, target infra.Target) error {
	// FIXME (wojciech): implement healthchecks instead of this hack
	time.Sleep(10 * time.Second)

	hermesHome := home + "/" + h.name
	config := hermesHome + "/config.toml"
	hermes := func(args ...string) *osexec.Cmd {
		return osexec.Command("hermes", append([]string{"--config", config}, args...)...)
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
id = '` + h.chainID1 + `'
rpc_addr = 'http://` + h.ip1 + `:26657'
grpc_addr = 'http://` + h.ip1 + `:9090'
websocket_addr = 'ws://` + h.ip1 + `:26657/websocket'
rpc_timeout = '10s'
account_prefix = 'sif'
key_name = '` + h.chainID1 + `'
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
id = '` + h.chainID2 + `'
rpc_addr = 'http://` + h.ip2 + `:26657'
grpc_addr = 'http://` + h.ip2 + `:9090'
websocket_addr = 'ws://` + h.ip2 + `:26657/websocket'
rpc_timeout = '10s'
account_prefix = 'sif'
key_name = '` + h.chainID2 + `'
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
	must.OK(ioutil.WriteFile(config, []byte(cfg), 0o600))

	err := exec.Run(ctx,
		hermes("keys", "add", h.chainID1, "--file", home+"/"+h.chainID1+".json"),
		hermes("keys", "add", h.chainID2, "--file", home+"/"+h.chainID2+".json"),
		hermes("create", "channel", h.chainID1, h.chainID2, "--port-a", "transfer", "--port-b", "transfer"),
	)
	if err != nil {
		return err
	}
	return target.DeployBinary(ctx, infra.Binary{
		Name: h.name,
		Path: "hermes",
		Args: []string{
			"--config", config,
			"start",
		},
	})
}

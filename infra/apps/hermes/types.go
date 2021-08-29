package hermes

import (
	"net"

	"github.com/wojciech-sif/localnet/infra"
)

// Peer is the interface required by hermes to be able to connect chains
type Peer interface {
	infra.HealthCheckCapable

	// ID returns chain id
	ID() string

	// IP returns ip used for connection
	IP() net.IP
}

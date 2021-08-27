package hermes

import "net"

// Peer is the interface required by hermes to be able to connect chains
type Peer interface {
	// ID returns chain id
	ID() string

	// IP returns ip used for connection
	IP() net.IP
}

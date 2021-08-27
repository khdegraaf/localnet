package infra

import "net"

// Config stores configuration
type Config struct {
	// EnvName is the name of created environment
	EnvName string

	// Target is the deployment target
	Target string

	// HomeDir is the path where all the files are kept
	HomeDir string

	// BinDir is the path where all binaries are present
	BinDir string

	// TMuxNetwork is the IP network for processes executed directly in tmux
	TMuxNetwork net.IP

	// TestingMode means we are in testing mode and deployment should not block execution
	TestingMode bool
}

package infra

import "net"

// Config stores configuration
type Config struct {
	// EnvName is the name of created environment
	EnvName string

	// HomeDir is the path where all the files are kept
	HomeDir string

	// TMuxStartIP is the starting IP for processes executed directly in tmux
	TMuxStartIP net.IP
}

package infra

import "net"

// Config stores configuration
type Config struct {
	// EnvName is the name of created environment
	EnvName string

	// SetName is the name of set
	SetName string

	// Target is the deployment target
	Target string

	// HomeDir is the path where all the files are kept
	HomeDir string

	// AppDir is the path where app data are stored
	AppDir string

	// LogDir is the path where logs are stored
	LogDir string

	// WrapperDir is the path where wrappers are stored
	WrapperDir string

	// BinDir is the path where all binaries are present
	BinDir string

	// TMuxNetwork is the IP network for processes executed directly in tmux
	TMuxNetwork net.IP

	// TestingMode means we are in testing mode and deployment should not block execution
	TestingMode bool

	// VerboseLogging turns on verbose logging
	VerboseLogging bool
}

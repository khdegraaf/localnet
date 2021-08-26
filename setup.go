package localnet

import (
	"net"
	"os"

	"github.com/ridge/must"
	"github.com/wojciech-malota-wojcik/ioc"
	"github.com/wojciech-sif/localnet/infra"
	"github.com/wojciech-sif/localnet/infra/targets"
)

// FIXME (wojciech): read it from CLI
const (
	envName = "localnet"
	target  = "tmux"
)

// IoC configures IoC container
func IoC(c *ioc.Container) {
	// FIXME (wojciech): read config from CLI
	c.Singleton(func() infra.Config {
		return infra.Config{
			EnvName:     envName,
			Target:      target,
			HomeDir:     must.String(os.UserHomeDir()) + "/.localnet/" + envName,
			TMuxStartIP: net.IPv4(127, 1, 0, 1), // 127.1.0.1
		}
	})
	c.TransientNamed("tmux", targets.NewTMux)
	c.TransientNamed("docker", targets.NewDocker)
	c.Transient(func(c *ioc.Container, config infra.Config) infra.Target {
		var target infra.Target
		c.ResolveNamed(config.Target, &target)
		return target
	})
}

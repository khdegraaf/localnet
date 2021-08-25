package exec

import (
	"context"
	"os"
	"os/exec"
	"syscall"

	"github.com/ridge/parallel"
	"github.com/wojciech-sif/localnet/lib/logger"
	"go.uber.org/zap"
)

// Run executes commands sequentially and terminates the running one gracefully if context is cancelled
func Run(ctx context.Context, cmds ...*exec.Cmd) error {
	log := logger.Get(ctx)

	for _, cmd := range cmds {
		cmd := cmd
		if cmd.Stdout == nil {
			cmd.Stdout = os.Stdout
		}
		if cmd.Stderr == nil {
			cmd.Stderr = os.Stderr
		}
		log.Info("Executing command", zap.String("cmd", cmd.String()))
		if err := cmd.Start(); err != nil {
			return err
		}

		err := parallel.Run(ctx, func(ctx context.Context, spawn parallel.SpawnFn) error {
			spawn("cmd", parallel.Exit, func(ctx context.Context) error {
				err := cmd.Wait()
				if ctx.Err() != nil {
					return ctx.Err()
				}
				return err
			})
			spawn("ctx", parallel.Exit, func(ctx context.Context) error {
				<-ctx.Done()
				_ = cmd.Process.Signal(syscall.SIGTERM)
				return ctx.Err()
			})
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// Tty returns handle to current tty
func Tty() *os.File {
	// For some reason /dev/stdin, /dev/stdout and /dev/stderr point to /proc/self/fd/0 which points to /dev/null.
	// That's why tmux doesn't work if os.Stdin is assigned to cmd.Stdin.
	// Workaround is to use /proc/self/fd/1 which points to correct tty
	tty, err := os.OpenFile("/proc/self/fd/1", os.O_RDWR, os.ModePerm)
	if err != nil {
		panic(err)
	}
	return tty
}

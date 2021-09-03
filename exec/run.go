package exec

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

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
		log.Debug("Executing command", zap.String("cmd", cmd.String()))
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

// Kill tries to terminate processes gracefully, after timeout it kills them
func Kill(ctx context.Context, pids []int) error {
	return parallel.Run(ctx, func(ctx context.Context, spawn parallel.SpawnFn) error {
		for _, pid := range pids {
			pid := pid
			spawn(fmt.Sprintf("%d", pid), parallel.Continue, func(ctx context.Context) error {
				return parallel.Run(ctx, func(ctx context.Context, spawn parallel.SpawnFn) error {
					proc, err := os.FindProcess(pid)
					if err != nil {
						return err
					}
					spawn("waiter", parallel.Exit, func(ctx context.Context) error {
						_, _ = proc.Wait()
						return nil
					})
					spawn("killer", parallel.Continue, func(ctx context.Context) error {
						if err := proc.Signal(syscall.SIGTERM); err != nil && !errors.Is(err, os.ErrProcessDone) {
							return err
						}
						select {
						case <-ctx.Done():
							return err
						case <-time.After(20 * time.Second):
						}
						if err := proc.Signal(syscall.SIGTERM); err != nil && !errors.Is(err, os.ErrProcessDone) {
							return err
						}
						return nil
					})
					return nil
				})
			})
		}
		return nil
	})
}

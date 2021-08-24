package exec

import (
	"context"
	"os"
	"os/exec"
	"syscall"

	"github.com/ridge/parallel"
)

// Run executes command and terminates it gracefully if context is cancelled
func Run(ctx context.Context, cmd *exec.Cmd) error {
	if cmd.Stdout == nil {
		cmd.Stdout = os.Stdout
	}
	if cmd.Stderr == nil {
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	return parallel.Run(ctx, func(ctx context.Context, spawn parallel.SpawnFn) error {
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
}

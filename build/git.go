package build

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	osexec "os/exec"

	"github.com/wojciech-sif/localnet/exec"
)

func gitFetch(ctx context.Context) error {
	return exec.Run(ctx, osexec.Command("git", "fetch", "-p"))
}

func gitStatusClean(ctx context.Context) error {
	buf := &bytes.Buffer{}
	cmd := osexec.Command("git", "status", "-s")
	cmd.Stdout = buf
	if err := exec.Run(ctx, cmd); err != nil {
		return err
	}
	if buf.Len() > 0 {
		fmt.Println("git status:")
		fmt.Println(buf)
		return errors.New("git status is not empty")
	}
	return nil
}

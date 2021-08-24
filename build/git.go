package build

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
)

func gitFetch(ctx context.Context) error {
	return runCmd(ctx, exec.Command("git", "fetch", "-p"))
}

func gitStatusClean(ctx context.Context) error {
	buf := &bytes.Buffer{}
	cmd := exec.Command("git", "status", "-s")
	cmd.Stdout = buf
	if err := runCmd(ctx, cmd); err != nil {
		return err
	}
	if buf.Len() > 0 {
		fmt.Println("git status:")
		fmt.Println(buf)
		return errors.New("git status is not empty")
	}
	return nil
}

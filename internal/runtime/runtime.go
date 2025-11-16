package runtime

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func Run(ctx context.Context, dir, command string, args ...string) error {
	cmd := exec.CommandContext(ctx, command, "-c", strings.Join(args, " "))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = dir
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("command '%s' failed: %w", command, err)
	}
	return nil
}

package runtime

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Run executes a command in an external shell of your choosing.
func Run(ctx context.Context, dir, shell string, args ...string) error {
	cmd := exec.CommandContext(ctx, shell, "-c", strings.Join(args, " "))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = dir
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("command '%s' failed: %w", shell, err)
	}
	return nil
}

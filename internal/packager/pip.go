package packager

import (
	"context"
	"github.com/djcass44/nib/cli/pkg/executor"
	"os"
	"path/filepath"
	"strings"
)

const commandPip = "pip"
const lockfilePip = "requirements.txt"

type Pip struct{}

// Detect checks to see if the build directory contains
// a Pip package list
func (*Pip) Detect(_ context.Context, bctx executor.BuildContext) bool {
	bctx.Logger.Process("Checking for Pip requirements.txt")

	_, err := os.Stat(filepath.Join(bctx.WorkingDir, lockfilePip))
	return err == nil
}

// Install installs packages using Pip
func (*Pip) Install(_ context.Context, bctx executor.BuildContext) error {
	bctx.Logger.Process("Executing install process")

	var extraArgs []string
	if val := os.Getenv(executor.EnvExtraArgs); val != "" {
		extraArgs = strings.Split(val, " ")
	}

	return executor.Exec(bctx, executor.Options{
		Command: commandPip,
		Args:    append([]string{"install", "-r", lockfilePip, "--cache-dir", bctx.CacheDir}, extraArgs...),
	})
}

// Build runs the Pip build script
func (*Pip) Build(_ context.Context, bctx executor.BuildContext) error {
	bctx.Logger.Process("Executing build process")
	bctx.Logger.Subprocess("Pip has no build process.")

	return nil
}

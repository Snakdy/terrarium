package packager

import (
	cbev1 "github.com/Snakdy/container-build-engine/pkg/api/v1"
	"github.com/Snakdy/container-build-engine/pkg/pipelines"
	"github.com/Snakdy/container-build-engine/pkg/pipelines/utils"
	"github.com/djcass44/nib/cli/pkg/executor"
	"github.com/go-logr/logr"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/scribe"
	"os"
)

const StatementPipInstall = "pip-install"

const commandPip = "pip"
const lockfilePip = "requirements.txt"

type PipInstall struct {
	options cbev1.Options
}

func (p *PipInstall) Run(ctx *pipelines.BuildContext) error {
	log := logr.FromContextOrDiscard(ctx.Context)
	log.V(7).Info("running statement pip install", "options", p.options)

	cacheDir, err := cbev1.GetRequired[string](p.options, "cache-dir")
	if err != nil {
		return err
	}
	installDir, err := cbev1.GetRequired[string](p.options, "install-dir")
	if err != nil {
		return err
	}

	buildContext := executor.BuildContext{
		WorkingDir: ctx.WorkingDirectory,
		CacheDir:   cacheDir,
		Clock:      chronos.DefaultClock,
		Logger:     scribe.NewLogger(os.Stdout),
	}

	return executor.Exec(buildContext, executor.Options{
		Command:  commandPip,
		Args:     []string{"install", "--ignore-installed", "-r", lockfilePip, "--cache-dir", cacheDir, "--target", installDir},
		ExtraEnv: ctx.ConfigFile.Config.Env,
	})
}

func (p *PipInstall) Name() string {
	return StatementPipInstall
}

func (p *PipInstall) SetOptions(options cbev1.Options) {
	if p.options == nil {
		p.options = map[string]any{}
	}
	utils.CopyMap(options, p.options)
}

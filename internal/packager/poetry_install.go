package packager

import (
	"context"
	cbev1 "github.com/Snakdy/container-build-engine/pkg/api/v1"
	"github.com/Snakdy/container-build-engine/pkg/pipelines"
	"github.com/Snakdy/container-build-engine/pkg/pipelines/utils"
	"github.com/djcass44/nib/cli/pkg/executor"
	"github.com/go-logr/logr"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/scribe"
	"os"
)

const StatementPoetryInstall = "poetry-export"

const commandSh = "sh"
const lockfilePoetry = "poetry.lock"

type PoetryExport struct {
	options cbev1.Options
}

func (p *PoetryExport) Run(ctx *pipelines.BuildContext) error {
	log := logr.FromContextOrDiscard(ctx.Context)
	log.V(7).Info("running statement poetry install", "options", p.options)

	cacheDir, err := cbev1.GetRequired[string](p.options, "cache-dir")
	if err != nil {
		return err
	}

	ok, err := p.Detect(ctx.Context, ctx.WorkingDirectory)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	buildContext := executor.BuildContext{
		WorkingDir: ctx.WorkingDirectory,
		CacheDir:   cacheDir,
		Clock:      chronos.DefaultClock,
		Logger:     scribe.NewLogger(os.Stdout),
	}

	return executor.Exec(buildContext, executor.Options{
		Command:  commandSh,
		Args:     []string{"-c", "poetry export --without-urls --format requirements.txt > requirements.txt"},
		ExtraEnv: ctx.ConfigFile.Config.Env,
	})
}

func (p *PoetryExport) Name() string {
	return StatementPoetryInstall
}

func (p *PoetryExport) SetOptions(options cbev1.Options) {
	if p.options == nil {
		p.options = map[string]any{}
	}
	utils.CopyMap(options, p.options)
}

func (p *PoetryExport) Detect(ctx context.Context, dir string) (bool, error) {
	return detectFile(ctx, dir, lockfilePoetry)
}

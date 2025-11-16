package packager

import (
	"context"

	cbev1 "github.com/Snakdy/container-build-engine/pkg/api/v1"
	"github.com/Snakdy/container-build-engine/pkg/pipelines"
	"github.com/Snakdy/container-build-engine/pkg/pipelines/utils"
	"github.com/Snakdy/terrarium/internal/runtime"
	"github.com/go-logr/logr"
)

const StatementPipInstall = "pip-install"

const commandPip = "pip"
const lockfilePip = "requirements.txt"

type PipInstall struct {
	options cbev1.Options
}

func (p *PipInstall) Run(ctx *pipelines.BuildContext, _ ...cbev1.Options) (cbev1.Options, error) {
	log := logr.FromContextOrDiscard(ctx.Context)
	log.V(7).Info("running statement pip install", "options", p.options)

	cacheDir, err := cbev1.GetRequired[string](p.options, "cache-dir")
	if err != nil {
		return cbev1.Options{}, err
	}
	installDir, err := cbev1.GetRequired[string](p.options, "install-dir")
	if err != nil {
		return cbev1.Options{}, err
	}

	err = runtime.Run(ctx.Context, ctx.WorkingDirectory, commandSh, commandPip+" install  --disable-pip-version-check --ignore-installed -r "+lockfilePip+" --cache-dir "+cacheDir+" --target "+installDir)
	if err != nil {
		log.Error(err, "script execution failed")
		return cbev1.Options{}, err
	}
	return cbev1.Options{}, nil
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

func (p *PipInstall) Detect(ctx context.Context, dir string) (bool, error) {
	return detectFile(ctx, dir, lockfilePip)
}

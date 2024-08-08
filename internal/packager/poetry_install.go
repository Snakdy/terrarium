package packager

import (
	"context"
	cbev1 "github.com/Snakdy/container-build-engine/pkg/api/v1"
	"github.com/Snakdy/container-build-engine/pkg/pipelines"
	"github.com/Snakdy/container-build-engine/pkg/pipelines/utils"
	"github.com/go-logr/logr"
	"os"
	"os/exec"
)

const StatementPoetryInstall = "poetry-export"

const commandSh = "sh"
const lockfilePoetry = "poetry.lock"

type PoetryExport struct {
	options cbev1.Options
}

func (p *PoetryExport) Run(ctx *pipelines.BuildContext, _ ...cbev1.Options) (cbev1.Options, error) {
	log := logr.FromContextOrDiscard(ctx.Context)
	log.V(7).Info("running statement poetry install", "options", p.options)

	ok, err := p.Detect(ctx.Context, ctx.WorkingDirectory)
	if err != nil {
		return cbev1.Options{}, err
	}
	if !ok {
		return cbev1.Options{}, nil
	}

	cmd := exec.CommandContext(ctx.Context, commandSh, "-c", "echo $PATH && poetry export --without-urls --format requirements.txt > requirements.txt")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = ctx.WorkingDirectory
	err = cmd.Run()
	if err != nil {
		log.Error(err, "script execution failed")
		return cbev1.Options{}, err
	}
	return cbev1.Options{}, nil
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

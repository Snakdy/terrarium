package pipinstall

import (
	"fmt"
	"os"
	"os/exec"

	cbev1 "github.com/Snakdy/container-build-engine/pkg/api/v1"
	"github.com/Snakdy/container-build-engine/pkg/pipelines"
	"github.com/Snakdy/container-build-engine/pkg/pipelines/utils"
	"github.com/go-logr/logr"
)

func (s *Statement) Run(ctx *pipelines.BuildContext, _ ...cbev1.Options) (cbev1.Options, error) {
	log := logr.FromContextOrDiscard(ctx.Context)
	log.V(7).Info("running statement pip install", "options", s.options)

	name, err := cbev1.GetRequired[string](s.options, "name")
	if err != nil {
		return cbev1.Options{}, err
	}
	enabled, err := cbev1.GetRequired[bool](s.options, "enabled")
	if err != nil {
		return cbev1.Options{}, err
	}
	if !enabled {
		return cbev1.Options{}, nil
	}

	cmd := exec.CommandContext(ctx.Context, "/bin/sh", "-c", fmt.Sprintf(`pip install "%s"`, name))
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

func (*Statement) Name() string {
	return Name
}

func (s *Statement) SetOptions(options cbev1.Options) {
	if s.options == nil {
		s.options = map[string]any{}
	}
	utils.CopyMap(options, s.options)
}

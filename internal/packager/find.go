package packager

import (
	"context"
	"github.com/Snakdy/container-build-engine/pkg/pipelines"
	"github.com/go-logr/logr"
	"os"
	"path/filepath"
)

var buildTools = []BuildTool{
	&PipInstall{},
}

type BuildTool interface {
	Detect(ctx context.Context, dir string) (bool, error)
}

func Detect(ctx context.Context, dir string) (pipelines.PipelineStatement, error) {
	log := logr.FromContextOrDiscard(ctx)
	for _, tool := range buildTools {
		ok, err := tool.Detect(ctx, dir)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		statement := tool.(pipelines.PipelineStatement)
		log.Info("detected build tool", "name", statement.Name())

		return statement, nil
	}
	log.Info("could not detect an external build tool - defaulting to pip")
	return buildTools[0].(pipelines.PipelineStatement), nil
}

func detectFile(ctx context.Context, dir, filename string) (bool, error) {
	log := logr.FromContextOrDiscard(ctx).WithValues("filename", filename)
	log.V(3).Info("checking for lockfile", "path", filepath.Join(dir, filename))

	_, err := os.Stat(filepath.Join(dir, filename))
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		log.Error(err, "failed to check for lockfile")
		return false, err
	}

	return true, nil
}

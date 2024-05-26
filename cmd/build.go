package cmd

import (
	"fmt"
	"github.com/Snakdy/container-build-engine/pkg/builder"
	"github.com/Snakdy/container-build-engine/pkg/containers"
	"github.com/Snakdy/container-build-engine/pkg/pipelines"
	"github.com/Snakdy/terrarium/internal/packager"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build and publish container images from the given directory.",
	Long:  "This sub-command builds the provided directory into static files, containerises them, and publishes them.",
	Args:  cobra.ExactArgs(1),
	RunE:  buildExec,
}

func init() {
	buildCmd.Flags().StringSliceP(flagTag, "t", []string{"latest"}, "tags to push")
	buildCmd.Flags().String(flagSave, "", "path to save the image as a tar archive")
	buildCmd.Flags().String(flagEntrypoint, "", "path to the Python file that will be executed")
	buildCmd.Flags().String(flagPlatform, "linux/amd64", "build platform")

	_ = buildCmd.MarkFlagRequired(flagEntrypoint)
	_ = buildCmd.MarkFlagFilename(flagEntrypoint, ".py")
}

func buildExec(cmd *cobra.Command, args []string) error {
	workingDir := args[0]
	localPath, _ := cmd.Flags().GetString(flagSave)
	cacheDir := os.Getenv(EnvCache)
	if cacheDir == "" {
		cacheDir = filepath.Join(os.TempDir(), ".terrarium-cache")
	}
	entrypoint, _ := cmd.Flags().GetString(flagEntrypoint)

	platform, _ := cmd.Flags().GetString(flagPlatform)
	imgPlatform, err := v1.ParsePlatform(platform)
	if err != nil {
		return err
	}

	installDir, err := os.MkdirTemp("", "pip-install-*")
	if err != nil {
		return err
	}

	pkgDir := filepath.Join(installDir, "packages")

	installStatement, err := packager.Detect(cmd.Context(), workingDir)
	if err != nil {
		return err
	}

	statements := []pipelines.OrderedPipelineStatement{
		{
			ID: "set-build-env",
			Options: map[string]any{
				"PYTHONUSERBASE": installDir,
				"PATH":           "${PATH}:" + filepath.Join(installDir, "bin") + ":" + os.Getenv("PATH"),
				"POETRY_VIRTUALENVS_OPTIONS_NO_SETUPTOOLS": "true",
				"POETRY_VIRTUALENVS_OPTIONS_NO_PIP":        "true",
				"POETRY_VIRTUALENVS_OPTIONS_ALWAYS_COPY":   "true",
				"POETRY_VIRTUALENVS_CREATE":                "false",
				"POETRY_CACHE_DIR":                         cacheDir,
			},
			Statement: &pipelines.Env{},
		},
		{
			ID: "poetry-export",
			Options: map[string]any{
				"cache-dir": cacheDir,
			},
			Statement: &packager.PoetryExport{},
			DependsOn: []string{"set-build-env"},
		},
		{
			ID: "pkg-install",
			Options: map[string]any{
				"cache-dir":   cacheDir,
				"install-dir": pkgDir,
			},
			Statement: installStatement,
			DependsOn: []string{"set-build-env", "poetry-export"},
		},
		{
			ID: "copy-python-packages",
			Options: map[string]any{
				"src": installDir,
				"dst": "/var/run/pip",
			},
			Statement: &pipelines.Dir{},
			DependsOn: []string{"pkg-install"},
		},
		{
			ID: "set-run-env",
			Options: map[string]any{
				"PYTHONUSERBASE": "/var/run/pip",
				"PATH":           "${PATH}:/var/run/pip/bin",
				"PYTHONPATH":     "/var/run/pip/packages:${PYTHONPATH}",
			},
			Statement: &pipelines.Env{},
			DependsOn: []string{"copy-python-packages"},
		},
		{
			ID: "copy-working-dir",
			Options: map[string]any{
				"src": workingDir,
				"dst": filepath.Join("${HOME}", "app"),
			},
			Statement: &pipelines.Dir{},
			DependsOn: []string{"set-run-env", "copy-python-packages"},
		},
	}

	// 4. add static files to base image
	baseImage := os.Getenv(EnvBaseImage)
	if baseImage == "" {
		baseImage = "python:3.12"
	}

	b, err := builder.NewBuilder(cmd.Context(), baseImage, statements, builder.Options{
		WorkingDir:      workingDir,
		Entrypoint:      []string{"python3"},
		Command:         []string{filepath.Join("${HOME}", "app", entrypoint)},
		ForceEntrypoint: true,
		Metadata: builder.MetadataOptions{
			CreatedBy: "terrarium",
		},
	})
	if err != nil {
		return err
	}
	img, err := b.Build(cmd.Context(), imgPlatform)
	if err != nil {
		return err
	}

	if localPath != "" {
		return containers.Save(cmd.Context(), img, "image", localPath)
	}
	tags, _ := cmd.Flags().GetStringSlice(flagTag)
	for _, tag := range tags {
		if err := containers.Push(cmd.Context(), img, fmt.Sprintf("%s:%s", os.Getenv(EnvDockerRepo), tag)); err != nil {
			return err
		}
	}

	return nil
}

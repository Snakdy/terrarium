package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"chainguard.dev/apko/pkg/apk/fs"
	"github.com/Snakdy/container-build-engine/pkg/builder"
	"github.com/Snakdy/container-build-engine/pkg/containers"
	"github.com/Snakdy/container-build-engine/pkg/pipelines"
	"github.com/Snakdy/terrarium/internal/packager"
	"github.com/Snakdy/terrarium/internal/pipinstall"
	"github.com/go-logr/logr"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/spf13/cobra"
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

	buildCmd.Flags().Bool(flagInstallPoetry, false, "(deprecated) whether to 'pip install poetry' before trying to use Poetry.")
	buildCmd.Flags().String(flagPoetryVersion, "", "if set, controls the version of Poetry installed")

	buildCmd.Flags().String(flagInstallTool, "", "name of the build tool to install. Can include a version (e.g., 'poetry<2.0.0')")

	_ = buildCmd.MarkFlagRequired(flagEntrypoint)
	_ = buildCmd.MarkFlagFilename(flagEntrypoint, ".py")
}

func buildExec(cmd *cobra.Command, args []string) error {
	log := logr.FromContextOrDiscard(cmd.Context())
	workingDir := args[0]
	localPath, _ := cmd.Flags().GetString(flagSave)
	cacheDir := os.Getenv(EnvCache)
	if cacheDir == "" {
		cacheDir = filepath.Join(os.TempDir(), ".terrarium-cache")
	}
	entrypoint, _ := cmd.Flags().GetString(flagEntrypoint)
	installPoetry, _ := cmd.Flags().GetBool(flagInstallPoetry)
	poetryVersion, _ := cmd.Flags().GetString(flagPoetryVersion)

	installTool, _ := cmd.Flags().GetString(flagInstallTool)
	// if the --install-poetry command has been
	// set, use that instead
	if installPoetry {
		log.Info("the '--install-poetry' flag is deprecated, please use '--install-tool=poetry' instead")
		installTool = "poetry==" + poetryVersion
		if poetryVersion == "" {
			installTool = "poetry<2.0.0"
		} else {
			log.Info("the '--poetry-version' flag is deprecated, please use '--install-tool=poetry<2.0.0' instead")
		}
	}

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
			ID: pipinstall.Name,
			Options: map[string]any{
				"enabled": installTool != "",
				"name":    installTool,
			},
			Statement: &pipinstall.Statement{},
			DependsOn: []string{"set-build-env"},
		},
		{
			ID: packager.StatementPoetryInstall,
			Options: map[string]any{
				"cache-dir": cacheDir,
			},
			Statement: &packager.PoetryExport{},
			DependsOn: []string{"set-build-env", pipinstall.Name},
		},
		{
			ID: packager.StatementUVSync,
			Options: map[string]any{
				"cache-dir": cacheDir,
			},
			Statement: &packager.UVSync{},
			DependsOn: []string{"set-build-env", pipinstall.Name},
		},
		{
			ID: "pkg-install",
			Options: map[string]any{
				"cache-dir":   cacheDir,
				"install-dir": pkgDir,
			},
			Statement: installStatement,
			DependsOn: []string{"set-build-env", packager.StatementPoetryInstall, packager.StatementUVSync},
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

	// 4. add static files to the base image
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
		FS: fs.NewMemFS(),
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

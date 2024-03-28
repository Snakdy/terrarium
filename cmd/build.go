package cmd

import (
	"fmt"
	"github.com/djcass44/all-your-base/pkg/containerutil"
	"github.com/djcass44/nib/cli/pkg/build"
	"github.com/djcass44/nib/cli/pkg/executor"
	"github.com/djcass44/terrarium/internal/packager"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/scribe"
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
}

var buildEngines = []executor.PackageManager{
	&packager.Pip{},
}

func buildExec(cmd *cobra.Command, args []string) error {
	workingDir := args[0]
	localPath, _ := cmd.Flags().GetString(flagSave)
	cacheDir := os.Getenv(EnvCache)
	if cacheDir == "" {
		cacheDir = filepath.Join(os.TempDir(), ".terrarium-cache")
	}

	bctx := executor.BuildContext{
		WorkingDir: workingDir,
		CacheDir:   cacheDir,
		Clock:      chronos.DefaultClock,
		Logger:     scribe.NewLogger(os.Stdout),
	}
	// 1. install dependencies
	pkg := buildEngines[0]
	for _, engine := range buildEngines {
		ok := engine.Detect(cmd.Context(), bctx)
		if ok {
			pkg = engine
			break
		}
	}
	err := pkg.Install(cmd.Context(), bctx)
	if err != nil {
		return err
	}

	// 2. build
	err = pkg.Build(cmd.Context(), bctx)
	if err != nil {
		return err
	}

	platform, err := v1.ParsePlatform("linux/amd64")
	if err != nil {
		return err
	}

	// 4. add static files to base image
	baseImage := os.Getenv(EnvBaseImage)
	if baseImage == "" {
		baseImage = "python:3.12"
	}
	options := build.Options{
		Author:      build.NibAuthor,
		ExtraEnv:    []string{"PYTHONUSERBASE=/var/run/pip"},
		Platform:    platform,
		EnvDataPath: build.NibDataPath,
	}
	img, err := build.Append(cmd.Context(), baseImage, options, build.LayerPath{
		Path:   workingDir,
		Chroot: build.DefaultChroot,
	}, build.LayerPath{
		Path:   filepath.Join(workingDir, ".pip"),
		Chroot: "/var/run/pip",
	})
	if err != nil {
		return err
	}
	if localPath != "" {
		return containerutil.Save(cmd.Context(), img, "image", localPath)
	}
	tags, _ := cmd.Flags().GetStringSlice(flagTag)
	for _, tag := range tags {
		if err := containerutil.Push(cmd.Context(), img, fmt.Sprintf("%s:%s", os.Getenv(EnvDockerRepo), tag)); err != nil {
			return err
		}
	}

	return nil
}

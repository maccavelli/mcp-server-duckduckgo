// Package main provides the canonical self-update command.
package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/maccavelli/mcplib/selfupdate"
	"github.com/spf13/cobra"
)

// skipConfigAnnotation marks a command that must not initialize configuration.
// The root pre-run honours it, which keeps `update` from creating the cache
// directory and config.yaml that config.Load writes, acquiring the serve
// singleton lock, or opening the BuntDB store (MADR 0005 F9).
const skipConfigAnnotation = "selfupdate.skip-config"

// skipConfigValue is the only annotation value that opts a command out.
const skipConfigValue = "true"

// releaseBuildKind is the only stamped value that marks a release build.
const releaseBuildKind = "release"

// updateRepository and updatePlatforms are this product's frozen release
// identity and matrix. A platform outside the set is rejected before any
// network call.
var (
	updateRepository = selfupdate.Repository{Owner: "maccavelli", Name: "mcp-server-duckduckgo"}
	updatePlatforms  = []selfupdate.Platform{
		{OS: "linux", Arch: "amd64"},
		{OS: "darwin", Arch: "arm64"},
		{OS: "windows", Arch: "amd64"},
	}
)

// updateOperationTimeout bounds one whole update.
const updateOperationTimeout = 15 * time.Minute

// newUpdater is the construction seam; tests replace it so the CLI matrix
// makes no live GitHub call.
var newUpdater = defaultUpdater

func defaultUpdater(errw io.Writer) (*selfupdate.Updater, error) {
	limits := selfupdate.DefaultLimits()
	source, err := selfupdate.NewGitHubSource(selfupdate.GitHubOptions{
		Repository: updateRepository,
		Client:     &http.Client{Timeout: updateOperationTimeout},
		UserAgent:  updateRepository.Name + "/" + RawVersion,
		Limits:     limits,
	})
	if err != nil {
		return nil, err
	}
	selector, err := selfupdate.NewExactAssetSelector(updatePlatforms)
	if err != nil {
		return nil, err
	}
	// Standalone only.
	installer, err := selfupdate.NewStandaloneInstaller(selfupdate.InstallOptions{})
	if err != nil {
		return nil, err
	}
	return selfupdate.New(selfupdate.Config{
		Source:    source,
		Versions:  selfupdate.NewStrictVersionPolicy(),
		Assets:    selector,
		Installer: installer,
		Reporter:  selfupdate.NewTextReporter(errw),
		Confirmer: selfupdate.NewTerminalConfirmer(os.Stdin, errw),
		Limits:    limits,
	})
}

// buildKind maps the linker stamp. Only the exact string "release" counts; a
// bool cannot be set with -X, so anything else is a local build.
func buildKind() selfupdate.BuildKind {
	if RawBuildKind == releaseBuildKind {
		return selfupdate.ReleaseBuild
	}
	return selfupdate.LocalBuild
}

func newUpdateCmd() *cobra.Command {
	var (
		check, yes, force bool
		targetVersion     string
	)
	cmd := &cobra.Command{
		Use:         "update",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{skipConfigAnnotation: skipConfigValue},
		Short:       "Download and install a GitHub release for " + updateRepository.Name,
		Long: `Check GitHub releases, verify the release checksum, and replace this
executable.

Exit codes: 0 = up to date or declined, 10 = --check found an actionable
target, 1 = error. Set GH_TOKEN or GITHUB_TOKEN to raise API rate limits.

This command creates no cache directory or config file and starts no datastore,
browser, search engine, tool registry, transport or MCP server.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Reporter and confirmation go to the command's error stream, so
			// the protocol stdout is never written by an update.
			updater, err := newUpdater(cmd.ErrOrStderr())
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), updateOperationTimeout)
			defer cancel()
			_, err = updater.Run(ctx, selfupdate.Request{
				Product:        updateRepository.Name,
				CurrentVersion: RawVersion,
				CurrentBuild:   buildKind(),
				TargetVersion:  targetVersion,
				CheckOnly:      check,
				Force:          force,
				Yes:            yes,
			})
			return err
		},
	}
	cmd.Flags().BoolVar(&check, "check", false, "report only; exit 0 up to date, 10 actionable, 1 error")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "approve the selected operation without prompting")
	cmd.Flags().BoolVar(&force, "force", false, "replace a local build, or reinstall the selected version")
	cmd.Flags().StringVar(&targetVersion, "version", "", "install this exact release tag (vX.Y.Z); a lower tag is an explicit rollback")
	return cmd
}

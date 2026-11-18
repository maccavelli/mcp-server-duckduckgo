// Package main provides the canonical self-update command.
package main

import (
	"context"
	"errors"
	"os"

	"github.com/maccavelli/go-selfupdate-lib/buildinfo"
	"github.com/maccavelli/go-selfupdate-lib/selfupdate"
	"github.com/maccavelli/go-selfupdate-lib/selfupdate/cli"
	"github.com/spf13/cobra"
)

// skipConfigAnnotation marks a command that must not initialize configuration.
// The root pre-run honours it, which keeps `update` from creating the cache
// directory and config.yaml that config.Load writes, acquiring the serve
// singleton lock, or opening the BuntDB store (MADR 0005 F9).
const skipConfigAnnotation = "selfupdate.skip-config"

// skipConfigValue is the only annotation value that opts a command out.
const skipConfigValue = "true"

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

// newUpdateUpdater is the construction seam. Tests replace it so the CLI
// matrix makes no live GitHub call.
var newUpdateUpdater = defaultNewUpdateUpdater

// updateOptions builds the streams for one run. Tests call it directly.
var updateOptions = commandUpdateOptions

func defaultNewUpdateUpdater() (*selfupdate.Updater, error) {
	src, err := selfupdate.NewGitHubSource(selfupdate.GitHubOptions{
		Repository: updateRepository,
		Client:     selfupdate.DefaultHTTPClient(),
		UserAgent:  selfupdate.UserAgent(updateRepository.Name, buildinfo.Identity().Current()),
		Limits:     selfupdate.DefaultLimits(),
	})
	if err != nil {
		return nil, err
	}
	selector, err := selfupdate.NewExactAssetSelector(updatePlatforms)
	if err != nil {
		return nil, err
	}
	installer, err := selfupdate.NewStandaloneInstaller(selfupdate.InstallOptions{})
	if err != nil {
		return nil, err
	}
	return selfupdate.New(selfupdate.Config{
		Source:    src,
		Versions:  selfupdate.NewStrictVersionPolicy(),
		Assets:    selector,
		Installer: installer,
		Reporter:  selfupdate.DiscardReporter(),
		Confirmer: selfupdate.NonInteractiveConfirmer(),
		Limits:    selfupdate.DefaultLimits(),
	})
}

func commandUpdateOptions(cmd *cobra.Command, asJSON bool) (cli.Options, error) {
	in := cmd.InOrStdin()
	o := cli.Options{Stderr: cmd.ErrOrStderr(), Stdin: in, JSON: asJSON}
	if in == os.Stdin {
		o.Interactive = cli.StdioOptions().Interactive
	}
	if asJSON {
		if protocolStdout == nil {
			return cli.Options{}, errors.New("update --json: protocol stdout is not preserved")
		}
		o.Stdout = protocolStdout
	}
	return o, nil
}

func newUpdateCmd() *cobra.Command {
	var flags cli.Flags
	cmd := &cobra.Command{
		Use:           "update",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		Annotations:   map[string]string{skipConfigAnnotation: skipConfigValue},
		Short:         "Download and install a GitHub release for " + updateRepository.Name,
		Long: `Check GitHub releases, verify the release checksum, and replace this
executable.

Exit codes: 0 = up to date or declined, 10 = --check found an actionable
target, 1 = error. Set GH_TOKEN or GITHUB_TOKEN to raise API rate limits.

This command creates no cache directory or config file and starts no datastore,
browser, search engine, tool registry, transport or MCP server.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			req, err := flags.Request(updateRepository.Name, buildinfo.Identity())
			if err != nil {
				return err
			}
			o, err := updateOptions(cmd, flags.JSON)
			if err != nil {
				return err
			}
			u, err := newUpdateUpdater()
			if err != nil {
				return err
			}
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			_, err = cli.Run(ctx, u, req, o)
			return err
		},
	}
	flags.Bind(cmd.Flags())
	return cmd
}

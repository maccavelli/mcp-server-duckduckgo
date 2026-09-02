package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/maccavelli/mcplib/selfupdate"
	"github.com/spf13/cobra"
)

// newTestRoot mirrors the real root wiring with the updater construction seam
// replaced, so no test makes a live GitHub call.
func newTestRoot(t *testing.T, build func(io.Writer) (*selfupdate.Updater, error)) (*cobra.Command, *bytes.Buffer) {
	t.Helper()
	prev := newUpdater
	newUpdater = build
	t.Cleanup(func() { newUpdater = prev })

	var out bytes.Buffer
	root := &cobra.Command{Use: "mcp-server-duckduckgo"}
	root.PersistentPreRunE = RootCmd.PersistentPreRunE
	root.AddCommand(newUpdateCmd())
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetIn(strings.NewReader(""))
	return root, &out
}

func unreachableBuild(io.Writer) (*selfupdate.Updater, error) {
	return nil, errors.New("updater construction should not have been reached")
}

// TestUpdateDoesNotLoadConfig is the sentinel this phase exists for.
// cobra.OnInitialize(config.Load) ran for every command, and config.Load
// creates the cache directory and writes config.yaml — so merely checking for
// an update left files on disk.
func TestUpdateDoesNotLoadConfig(t *testing.T) {
	loaded := 0
	prev := loadConfig
	loadConfig = func() { loaded++ }
	t.Cleanup(func() { loadConfig = prev })

	root, _ := newTestRoot(t, unreachableBuild)
	root.SetArgs([]string{"update", "--check"})
	_ = root.ExecuteContext(context.Background())
	if loaded != 0 {
		t.Fatalf("update loaded configuration %d times; it must not create the cache directory or config.yaml", loaded)
	}
}

// TestUpdateLeavesNoCacheDirectory is the observable form of the same
// guarantee: point HOME and XDG_CACHE_HOME at an empty temporary directory and
// require that a check writes nothing into it.
func TestUpdateLeavesNoCacheDirectory(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CACHE_HOME", filepath.Join(tmp, "cache"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, "config"))

	root, _ := newTestRoot(t, unreachableBuild)
	root.SetArgs([]string{"update", "--check"})
	_ = root.ExecuteContext(context.Background())

	entries, err := os.ReadDir(tmp)
	if err != nil {
		t.Fatalf("read temp home: %v", err)
	}
	if len(entries) != 0 {
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Fatalf("update created %v under an empty home; it must create nothing", names)
	}
}

// TestOrdinaryCommandStillLoadsConfig proves the opt-out is scoped to the
// annotation and did not disable configuration for everything else.
func TestOrdinaryCommandStillLoadsConfig(t *testing.T) {
	loaded := 0
	prev := loadConfig
	loadConfig = func() { loaded++ }
	t.Cleanup(func() { loadConfig = prev })

	root := &cobra.Command{Use: "mcp-server-duckduckgo"}
	root.PersistentPreRunE = RootCmd.PersistentPreRunE
	ran := false
	root.AddCommand(&cobra.Command{Use: "ordinary", RunE: func(*cobra.Command, []string) error { ran = true; return nil }})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"ordinary"})
	if err := root.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("ordinary command failed: %v", err)
	}
	if !ran {
		t.Fatal("ordinary command did not run")
	}
	if loaded != 1 {
		t.Fatalf("ordinary command loaded configuration %d times, want 1; the opt-out is too wide", loaded)
	}
}

// TestUpdateWritesNothingToProtocolStdout proves update output is routed to the
// command's error stream. Recall's stdout carries JSON-RPC.
func TestUpdateWritesNothingToProtocolStdout(t *testing.T) {
	var stdout, stderr bytes.Buffer
	prev := newUpdater
	newUpdater = func(w io.Writer) (*selfupdate.Updater, error) {
		if w != io.Writer(&stderr) {
			t.Error("updater was given a stream other than the command's error stream")
		}
		return nil, errors.New("stop before any work")
	}
	t.Cleanup(func() { newUpdater = prev })

	cmd := newUpdateCmd()
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetIn(strings.NewReader(""))
	cmd.SetArgs([]string{"--check"})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	_ = cmd.ExecuteContext(context.Background())
	if stdout.Len() != 0 {
		t.Fatalf("update wrote %q to protocol stdout", stdout.String())
	}
}

func TestUpdateCommandIsAnnotatedToSkipConfig(t *testing.T) {
	if got := newUpdateCmd().Annotations[skipConfigAnnotation]; got != skipConfigValue {
		t.Fatalf("annotation = %q, want %q", got, skipConfigValue)
	}
}

func TestUpdateFlagSurface(t *testing.T) {
	cmd := newUpdateCmd()
	for _, name := range []string{"check", "yes", "force", "version"} {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("missing --%s", name)
		}
	}
	if f := cmd.Flags().ShorthandLookup("y"); f == nil || f.Name != "yes" {
		t.Error("missing -y shorthand for --yes")
	}
}

func TestUpdateRejectsPositionalArgs(t *testing.T) {
	cmd := newUpdateCmd()
	if err := cmd.Args(cmd, []string{"stray"}); err == nil {
		t.Fatal("expected positional arguments to be rejected")
	}
}

func TestUpdateRejectsContradictoryFlags(t *testing.T) {
	for _, flag := range []string{"--yes", "--force"} {
		t.Run(flag, func(t *testing.T) {
			root, _ := newTestRoot(t, defaultUpdater)
			root.SetArgs([]string{"update", "--check", flag})
			err := root.ExecuteContext(context.Background())
			if err == nil {
				t.Fatalf("--check %s was accepted", flag)
			}
			if errors.Is(err, selfupdate.ErrUpdateAvailable) {
				t.Fatal("contradiction was not detected before evaluation")
			}
		})
	}
}

func TestUpdateUsesCallerContext(t *testing.T) {
	root, _ := newTestRoot(t, defaultUpdater)
	root.SetArgs([]string{"update", "--check"})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := root.ExecuteContext(ctx); err == nil {
		t.Fatal("a cancelled caller context did not abort the command")
	}
}

func TestBuildKindMapping(t *testing.T) {
	prev := RawBuildKind
	t.Cleanup(func() { RawBuildKind = prev })
	for _, tc := range []struct {
		stamp string
		want  selfupdate.BuildKind
	}{
		{"release", selfupdate.ReleaseBuild},
		{"local", selfupdate.LocalBuild},
		{"", selfupdate.LocalBuild},
		{"Release", selfupdate.LocalBuild},
	} {
		RawBuildKind = tc.stamp
		if got := buildKind(); got != tc.want {
			t.Errorf("RawBuildKind=%q -> %v, want %v", tc.stamp, got, tc.want)
		}
	}
}

// TestVersionStampIsEffective guards the defect this phase fixed: the Makefile
// passed -X main.RawVersion but no such variable existed, so the linker
// silently did nothing and this server had no version identity at all.
func TestVersionStampIsEffective(t *testing.T) {
	if Version != strings.TrimPrefix(RawVersion, "v") {
		t.Fatalf("Version %q is not derived from RawVersion %q", Version, RawVersion)
	}
	if RawBuildKind == releaseBuildKind {
		t.Fatal("an unstamped build must not claim to be a release build")
	}
	if err := selfupdate.NewStrictVersionPolicy().Validate(RawVersion); err == nil {
		t.Fatalf("default RawVersion %q validates as a release tag; it must not", RawVersion)
	}
}

func TestExitCodeMapping(t *testing.T) {
	if got := selfupdate.ExitCode(selfupdate.Result{}, nil); got != 0 {
		t.Errorf("nil -> %d, want 0", got)
	}
	if got := selfupdate.ExitCode(selfupdate.Result{}, selfupdate.ErrUpdateAvailable); got != 10 {
		t.Errorf("ErrUpdateAvailable -> %d, want 10", got)
	}
	if got := selfupdate.ExitCode(selfupdate.Result{}, errors.New("boom")); got != 1 {
		t.Errorf("generic -> %d, want 1", got)
	}
}

func TestUpdatePlatformsAreTheFrozenMatrix(t *testing.T) {
	want := map[string]bool{"linux/amd64": true, "darwin/arm64": true, "windows/amd64": true}
	if len(updatePlatforms) != len(want) {
		t.Fatalf("updatePlatforms = %v", updatePlatforms)
	}
	for _, p := range updatePlatforms {
		if !want[p.OS+"/"+p.Arch] {
			t.Errorf("unexpected platform %v", p)
		}
	}
	if _, err := selfupdate.NewExactAssetSelector(updatePlatforms); err != nil {
		t.Fatalf("selector rejected the frozen matrix: %v", err)
	}
}

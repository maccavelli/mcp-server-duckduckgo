package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/maccavelli/go-selfupdate-lib/selfupdate"
	"github.com/maccavelli/go-selfupdate-lib/selfupdate/selfupdatetest"
	"github.com/spf13/cobra"
)

// newTestRoot mirrors the real root wiring with the updater construction seam
// replaced, so no test makes a live GitHub call. It does not call Execute.
func newTestRoot(t *testing.T, build func() (*selfupdate.Updater, error)) (*cobra.Command, *bytes.Buffer) {
	t.Helper()
	prev := newUpdateUpdater
	newUpdateUpdater = build
	t.Cleanup(func() { newUpdateUpdater = prev })

	var out bytes.Buffer
	root := &cobra.Command{Use: "mcp-server-duckduckgo"}
	root.PersistentPreRunE = RootCmd.PersistentPreRunE
	root.AddCommand(newUpdateCmd())
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetIn(strings.NewReader(""))
	return root, &out
}

func unreachableBuild() (*selfupdate.Updater, error) {
	return nil, errors.New("updater construction should not have been reached")
}

func recordingBuild(calls *int) func() (*selfupdate.Updater, error) {
	return func() (*selfupdate.Updater, error) {
		*calls++
		return nil, errors.New("updater construction should not have been reached")
	}
}

// TestUpdateDoesNotLoadConfig is the sentinel this phase exists for.
// cobra.OnInitialize(config.Load) ran for every command, and config.Load
// creates the cache directory and writes config.yaml, so merely checking for
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

// TestUpdateWritesNothingToProtocolStdout proves --check keeps the library
// off the protocol stdout. The command output and error streams are different
// buffers.
func TestUpdateWritesNothingToProtocolStdout(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := newUpdateCmd()
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--check"})
	o, err := commandUpdateOptions(cmd, false)
	if err != nil {
		t.Fatalf("options: %v", err)
	}
	if o.Stdout != nil {
		t.Fatal("check Stdout is set; protocol stdout must stay unused")
	}
	if o.Stderr != cmd.ErrOrStderr() {
		t.Fatal("check Stderr is not the command error stream")
	}
	if stdout.Len() != 0 {
		t.Fatalf("update wrote %q to the command output buffer", stdout.String())
	}
}

func TestUpdateCommandIsAnnotatedToSkipConfig(t *testing.T) {
	if got := newUpdateCmd().Annotations[skipConfigAnnotation]; got != skipConfigValue {
		t.Fatalf("annotation = %q, want %q", got, skipConfigValue)
	}
}

func TestUpdateFlagSurface(t *testing.T) {
	cmd := newUpdateCmd()
	for _, name := range []string{"check", "yes", "force", "version", "dry-run", "json", "channel"} {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("missing --%s", name)
		}
	}
	if f := cmd.Flags().ShorthandLookup("y"); f == nil || f.Name != "yes" {
		t.Error("missing -y shorthand for --yes")
	}
	if !cmd.SilenceUsage || !cmd.SilenceErrors {
		t.Fatalf("SilenceUsage=%v SilenceErrors=%v", cmd.SilenceUsage, cmd.SilenceErrors)
	}
}

func TestUpdateRejectsPositionalArgs(t *testing.T) {
	cmd := newUpdateCmd()
	if err := cmd.Args(cmd, []string{"stray"}); err == nil {
		t.Fatal("expected positional arguments to be rejected")
	}
}

func TestUpdateRejectsContradictoryFlags(t *testing.T) {
	for _, flag := range []string{"--yes", "--force", "--dry-run"} {
		t.Run(flag, func(t *testing.T) {
			calls := 0
			root, _ := newTestRoot(t, recordingBuild(&calls))
			root.SetArgs([]string{"update", "--check", flag})
			err := root.ExecuteContext(context.Background())
			if err == nil {
				t.Fatalf("--check %s was accepted", flag)
			}
			if errors.Is(err, selfupdate.ErrUpdateAvailable) {
				t.Fatal("contradiction was not detected before evaluation")
			}
			if calls != 0 {
				t.Fatalf("constructor calls = %d, want 0", calls)
			}
		})
	}
}

func TestUpdateUsesCallerContext(t *testing.T) {
	if !reexecUnderHome(t) {
		return
	}
	root, _ := newTestRoot(t, defaultNewUpdateUpdater)
	root.SetArgs([]string{"update", "--check"})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := root.ExecuteContext(ctx)
	if err == nil {
		t.Fatal("a cancelled caller context did not abort the command")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled context returned %v; want an error that errors.Is reports as context.Canceled", err)
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

func TestJSONRequiresPreservedStdout(t *testing.T) {
	prev := protocolStdout
	protocolStdout = nil
	t.Cleanup(func() { protocolStdout = prev })

	calls := 0
	root, _ := newTestRoot(t, recordingBuild(&calls))
	root.SetArgs([]string{"update", "--json"})
	err := root.ExecuteContext(context.Background())
	if err == nil || err.Error() != "update --json: protocol stdout is not preserved" {
		t.Fatalf("err = %v", err)
	}
	if calls != 0 {
		t.Fatalf("constructor calls = %d, want 0", calls)
	}
}

func TestCheckAndJSONStreams(t *testing.T) {
	if !runningPlatformIsFrozen() {
		t.Skip("runtime platform is outside updatePlatforms")
	}
	if !reexecUnderHome(t) {
		return
	}
	before, err := os.Executable()
	if err != nil {
		t.Fatalf("executable: %v", err)
	}
	prevUpdater := newUpdateUpdater
	newUpdateUpdater = fakeUpdateUpdater
	prevProtocol := protocolStdout
	t.Cleanup(func() {
		newUpdateUpdater = prevUpdater
		protocolStdout = prevProtocol
	})

	var stdout, stderr, protocol bytes.Buffer
	run := func(args ...string) error {
		t.Helper()
		stdout.Reset()
		stderr.Reset()
		cmd := newUpdateCmd()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetIn(strings.NewReader(""))
		cmd.SetArgs(args)
		return cmd.ExecuteContext(context.Background())
	}

	protocolStdout = nil
	err = run("--check")
	if !errors.Is(err, selfupdate.ErrUpdateAvailable) {
		t.Fatalf("--check err = %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("--check wrote %q to the command output", stdout.String())
	}
	if !strings.Contains(stderr.String(), "mcp-server-duckduckgo: update available:") || !strings.Contains(stderr.String(), "local build: apply requires --force") {
		t.Fatalf("--check error stream = %q", stderr.String())
	}

	protocol.Reset()
	protocolStdout = &protocol
	err = run("--check", "--json")
	if !errors.Is(err, selfupdate.ErrUpdateAvailable) {
		t.Fatalf("--check --json err = %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("--check --json wrote %q to the command output", stdout.String())
	}
	if !strings.Contains(protocol.String(), `"kind":"result"`) || !strings.Contains(protocol.String(), `"exit_code":10`) {
		t.Fatalf("protocol stdout = %q", protocol.String())
	}
	after, err := os.Executable()
	if err != nil {
		t.Fatalf("executable after: %v", err)
	}
	if after != before {
		t.Fatalf("process executable changed from %s to %s", before, after)
	}
}

func fakeUpdateUpdater() (*selfupdate.Updater, error) {
	release := selfupdatetest.NewRelease("mcp-server-duckduckgo", "v1.2.3", updatePlatforms, func(selfupdate.Platform) []byte {
		return []byte("body")
	})
	source := selfupdatetest.NewFakeSource("v1.2.3", release)
	selector, err := selfupdate.NewExactAssetSelector(updatePlatforms)
	if err != nil {
		return nil, err
	}
	installer, err := selfupdate.NewStandaloneInstaller(selfupdate.InstallOptions{})
	if err != nil {
		return nil, err
	}
	return selfupdate.New(selfupdate.Config{
		Source:    source,
		Versions:  selfupdate.NewStrictVersionPolicy(),
		Assets:    selector,
		Installer: installer,
		Reporter:  selfupdate.DiscardReporter(),
		Confirmer: selfupdate.NonInteractiveConfirmer(),
		Limits:    selfupdate.DefaultLimits(),
	})
}

func runningPlatformIsFrozen() bool {
	for _, p := range updatePlatforms {
		if p.OS == runtime.GOOS && p.Arch == runtime.GOARCH {
			return true
		}
	}
	return false
}

// reexecUnderHomeEnv marks the process that was started from a copy of the
// test binary under the home directory.
const reexecUnderHomeEnv = "MCP_SERVER_DUCKDUCKGO_REEXEC_UNDER_HOME"

// reexecUnderHome runs t from a copy of this executable under the home
// directory. Empty InstallOptions allow only that directory, and a go test
// binary otherwise lives in the system temporary directory. It returns true
// when this process should run the test body.
func reexecUnderHome(t *testing.T) bool {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("executable: %v", err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("home directory: %v", err)
	}
	exeResolved, err := filepath.EvalSymlinks(exe)
	if err != nil {
		t.Fatalf("executable: %v", err)
	}
	homeResolved, err := filepath.EvalSymlinks(home)
	if err != nil {
		t.Fatalf("home directory: %v", err)
	}
	if pathUnder(exeResolved, homeResolved) {
		return true
	}
	if os.Getenv(reexecUnderHomeEnv) == "1" {
		t.Fatal("re-exec copy is still outside the home directory")
	}
	dir, err := os.MkdirTemp(homeResolved, ".mcp-server-duckduckgo-reexec-")
	if err != nil {
		t.Fatalf("re-exec directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Errorf("remove re-exec directory: %v", err)
		}
	})
	dst := filepath.Join(dir, "mcp-server-duckduckgo.test")
	if err := copyExecutable(exe, dst); err != nil {
		t.Fatalf("copy test binary: %v", err)
	}
	dstResolved, err := filepath.EvalSymlinks(dst)
	if err != nil {
		t.Fatalf("re-exec binary: %v", err)
	}
	if !pathUnder(dstResolved, homeResolved) {
		t.Fatal("re-exec copy is still outside the home directory")
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, dst, "-test.run", "^"+regexp.QuoteMeta(t.Name())+"$", "-test.count=1", "-test.v")
	cmd.Dir = wd
	cmd.Env = append(os.Environ(), reexecUnderHomeEnv+"=1")
	out, err := cmd.CombinedOutput()
	pass := "--- PASS: " + t.Name()
	if err != nil || !bytes.Contains(out, []byte(pass)) || bytes.Contains(out, []byte("--- FAIL:")) || bytes.Contains(out, []byte("--- SKIP:")) {
		t.Fatalf("re-exec under home: %v\n%s", err, out)
	}
	return false
}

func pathUnder(path, root string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func copyExecutable(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Chmod(dst, 0o755)
}

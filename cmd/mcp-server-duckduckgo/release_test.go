package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/maccavelli/go-selfupdate-lib/buildinfo"
	"github.com/maccavelli/go-selfupdate-lib/selfupdate"
	"github.com/maccavelli/go-selfupdate-lib/selfupdate/releasespec"
	"github.com/spf13/cobra"
)

func TestUnstampedIdentityIsLocal(t *testing.T) {
	id := buildinfo.Identity()
	if id.Kind != buildinfo.KindLocal {
		t.Fatalf("Kind = %s, want local", id.Kind)
	}
	if id.Current() != "dev" {
		t.Fatalf("Current() = %q, want dev", id.Current())
	}
	if err := selfupdate.NewStrictVersionPolicy().Validate(id.Current()); err == nil {
		t.Fatal("unstamped Current() validated as a release tag")
	}
}

func TestStampCase(t *testing.T) {
	id := buildinfo.Identity()
	switch os.Getenv("SELFUPDATE_STAMP_CASE") {
	case "":
		t.Skip("SELFUPDATE_STAMP_CASE is empty")
	case "release":
		if id.Kind != buildinfo.KindRelease || id.Current() != "v9.9.9" {
			t.Fatalf("release stamp: kind %s current %q", id.Kind, id.Current())
		}
	case "dirty":
		if id.Kind != buildinfo.KindLocal || id.Reason == "" {
			t.Fatalf("dirty stamp: kind %s reason %q", id.Kind, id.Reason)
		}
	case "capital":
		if id.Kind != buildinfo.KindLocal {
			t.Fatalf("capital stamp: kind %s", id.Kind)
		}
	case "local-tag":
		if id.Kind != buildinfo.KindLocal || id.Current() != "v9.9.9" {
			t.Fatalf("local-tag stamp: kind %s current %q", id.Kind, id.Current())
		}
	default:
		t.Fatalf("unknown SELFUPDATE_STAMP_CASE %q", os.Getenv("SELFUPDATE_STAMP_CASE"))
	}
}

func TestMakefileStamps(t *testing.T) {
	out := runMake(t, moduleRoot(t), []string{"-n", "build", "linux", "darwin-arm64", "windows-amd64"}, "v9.9.9", "release")
	if strings.Count(out, "buildinfo.version=") != 4 || strings.Count(out, "buildinfo.kind=") != 4 {
		t.Fatalf("stamp count version=%d kind=%d\n%s", strings.Count(out, "buildinfo.version="), strings.Count(out, "buildinfo.kind="), out)
	}
	if strings.Contains(out, "main.RawVersion") || strings.Contains(out, "main.RawBuildKind") {
		t.Fatalf("old stamp names remain:\n%s", out)
	}
	var buildLine, linuxLine string
	for _, line := range strings.Split(out, "\n") {
		if !strings.Contains(line, "buildinfo.version=") {
			continue
		}
		if strings.Contains(line, "GOOS=linux") {
			linuxLine = line
			continue
		}
		if strings.Contains(line, "-tags netgo") {
			buildLine = line
		}
	}
	for _, line := range []string{buildLine, linuxLine} {
		if !strings.Contains(line, "-tags netgo") || !strings.Contains(line, "-extldflags '-static'") {
			t.Fatalf("static flags missing: %q\n%s", line, out)
		}
	}
}

func TestIdentityPrintsBuildinfo(t *testing.T) {
	cmd := newIdentityCmd()
	if cmd.Annotations[skipConfigAnnotation] != skipConfigValue {
		t.Fatalf("annotation = %q", cmd.Annotations[skipConfigAnnotation])
	}
	if !cmd.SilenceUsage || !cmd.SilenceErrors {
		t.Fatalf("SilenceUsage=%v SilenceErrors=%v", cmd.SilenceUsage, cmd.SilenceErrors)
	}

	loaded := 0
	prevLoad := loadConfig
	loadConfig = func() { loaded++ }
	t.Cleanup(func() { loadConfig = prevLoad })

	prevStdout := OriginalStdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	OriginalStdout = w
	t.Cleanup(func() {
		OriginalStdout = prevStdout
		r.Close()
		w.Close()
	})

	var out bytes.Buffer
	root := &cobra.Command{Use: "mcp-server-duckduckgo"}
	root.PersistentPreRunE = RootCmd.PersistentPreRunE
	root.AddCommand(newIdentityCmd())
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"identity"})
	if err := root.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("identity: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close pipe: %v", err)
	}
	got, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read pipe: %v", err)
	}
	want := buildinfo.Identity().String() + "\n"
	if string(got) != want {
		t.Fatalf("identity wrote %q, want %q", got, want)
	}
	if out.Len() != 0 {
		t.Fatalf("command output = %q", out.String())
	}
	if loaded != 0 {
		t.Fatalf("identity loaded configuration %d times", loaded)
	}

	OriginalStdout = nil
	err = newIdentityCmd().ExecuteContext(context.Background())
	if err == nil || err.Error() != "identity: protocol stdout is not preserved" {
		t.Fatalf("nil stdout err = %v", err)
	}

	found := false
	for _, c := range RootCmd.Commands() {
		if c.Use == "identity" {
			found = true
		}
	}
	if !found {
		t.Fatal("RootCmd does not list identity")
	}
}

func TestReleaseSpec(t *testing.T) {
	root := moduleRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "selfupdate-release.json"))
	if err != nil {
		t.Fatalf("read spec: %v", err)
	}
	spec, err := releasespec.Parse(data)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(spec.Products) != 1 {
		t.Fatalf("products = %d", len(spec.Products))
	}
	p := spec.Products[0]
	if p.Name != "mcp-server-duckduckgo" || p.Package != "./cmd/mcp-server-duckduckgo" || strings.Join(p.IdentityArgs, " ") != "identity" {
		t.Fatalf("product = %+v", p)
	}
	if len(p.Tags) != 0 {
		t.Fatalf("tags = %v", p.Tags)
	}
	wantPlatforms := []releasespec.Platform{{OS: "linux", Arch: "amd64"}, {OS: "darwin", Arch: "arm64"}, {OS: "windows", Arch: "amd64"}}
	if len(spec.Platforms) != len(wantPlatforms) {
		t.Fatalf("platforms = %+v", spec.Platforms)
	}
	for i, p := range wantPlatforms {
		if spec.Platforms[i] != p {
			t.Fatalf("platform %d = %+v", i, spec.Platforms[i])
		}
	}
	if spec.Packaging != releasespec.PackagingBinary {
		t.Fatalf("packaging = %q", spec.Packaging)
	}
	if spec.InstallerName("mcp-server-duckduckgo") != "mcp-server-duckduckgo" {
		t.Fatalf("installer name = %q", spec.InstallerName("mcp-server-duckduckgo"))
	}
	prefix, err := spec.InstallerEnvPrefix("mcp-server-duckduckgo")
	if err != nil || prefix != "MCP_SERVER_DUCKDUCKGO" {
		t.Fatalf("env prefix = %q, %v", prefix, err)
	}
	scripts := spec.InstallerScripts()
	if len(scripts) != 2 || scripts[0] != "install.sh" || scripts[1] != "install.ps1" {
		t.Fatalf("scripts = %v", scripts)
	}
	names, err := spec.ExtraNames("v9.9.9")
	if err != nil || len(names) != 2 || names[0] != "install.sh" || names[1] != "install.ps1" {
		t.Fatalf("extra names = %v, %v", names, err)
	}
	if spec.Installer == nil || len(spec.Installer.Hooks) != 0 || len(spec.Extras) != 0 || len(spec.PrereleaseChannels) != 0 {
		t.Fatalf("installer=%+v extras=%v channels=%v", spec.Installer, spec.Extras, spec.PrereleaseChannels)
	}
}

func TestReleaseWorkflowPin(t *testing.T) {
	text := string(mustRead(t, filepath.Join(moduleRoot(t), ".github", "workflows", "ci.yml")))
	for _, needle := range []string{
		"maccavelli/go-selfupdate-lib/.github/workflows/build-selfupdate-release.yml@1dffdc9dee24a7534513dc51c9039408be37c09b",
		"maccavelli/go-selfupdate-lib/.github/workflows/publish-selfupdate-release.yml@1dffdc9dee24a7534513dc51c9039408be37c09b",
		"spec-path: selfupdate-release.json",
		"needs.build.outputs.rehearsal == 'false'",
		"needs.build.outputs.artifact-name",
		"needs.build.outputs.products-json",
		"needs.build.outputs.platforms-json",
		"needs.build.outputs.extra-assets-json",
		"needs.build.outputs.prerelease-channels-json",
		"https://github.com/maccavelli/go-shared.git",
		"fetch --depth 1 origin",
		"checkout --detach FETCH_HEAD",
		"golangci-lint@v2.15.0",
		`GOLANGCI_LINT="$(go env GOPATH)/bin/golangci-lint"`,
		"${{ runner.temp }}/go-shared",
		"make verify GO_SHARED=",
	} {
		if !strings.Contains(text, needle) {
			t.Errorf("workflow missing %s", needle)
		}
	}
	sha := goSharedSHA(t, text)
	for _, banned := range []string{
		"1dffdc9dee24a7534513dc51c9039408be37c09b",
		"10316bca82c00aeefde4bc614d183d343ed5818e",
		"8d4d89ff34ff79c1ed65f223cc3d83b3ec0550aa",
	} {
		if sha == banned {
			t.Errorf("GO_SHARED_SHA is %s", banned)
		}
	}
	for _, banned := range []string{
		"path: ${{ runner.temp }}/go-shared",
		"repository: maccavelli/go-shared",
		"bridge-release",
		"8d4d89ff34ff79c1ed65f223cc3d83b3ec0550aa",
		"make build-all",
		"extra-assets-json: '[]'",
		"go test ./...",
		"go vet ./...",
	} {
		if strings.Contains(text, banned) {
			t.Errorf("workflow contains %s", banned)
		}
	}
}

func TestMakefileVerify(t *testing.T) {
	root := moduleRoot(t)
	text := string(mustRead(t, filepath.Join(root, "Makefile")))
	var verifyLine string
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, "verify:") {
			verifyLine = line
		}
	}
	const want = "verify: mod-check fmt-check lint vet test vuln check-records markdownlint disclosure-check ## Runs the local and CI quality checks"
	if verifyLine != want {
		t.Fatalf("verify line = %q", verifyLine)
	}
	for _, banned := range []string{"go fmt", "actionlint", "check_ci_filters.py", "build-all", "-mode=binary"} {
		if strings.Contains(verifyLine, banned) {
			t.Errorf("verify recipe contains %s", banned)
		}
	}
	out := runMake(t, root, []string{"-n", "verify", "GO_SHARED=go-shared-under-test"}, "", "")
	for _, needle := range []string{
		`python3 "go-shared-under-test/scripts/check_disclosure.py" --repo .`,
		"gofmt -l",
		"go mod tidy -diff",
		"go mod verify",
		"go run golang.org/x/vuln/cmd/govulncheck@v1.10.0 ./...",
	} {
		if !strings.Contains(out, needle) {
			t.Errorf("make -n verify missing %s", needle)
		}
	}
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		if err == nil && strings.Contains(string(data), "module github.com/maccavelli/mcp-server-duckduckgo") {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("module root not found")
		}
		dir = parent
	}
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return data
}

func goSharedSHA(t *testing.T, workflow string) string {
	t.Helper()
	re := regexp.MustCompile(`(?m)^[ \t]*GO_SHARED_SHA:[ \t]*([0-9a-f]{40})[ \t]*$`)
	m := re.FindStringSubmatch(workflow)
	if m == nil {
		t.Fatal("GO_SHARED_SHA is not 40 lowercase hexadecimal characters")
	}
	return m[1]
}

func runMake(t *testing.T, root string, args []string, version, kind string) string {
	t.Helper()
	cmd := exec.Command("make", args...)
	cmd.Dir = root
	cmd.Env = os.Environ()
	if version != "" {
		cmd.Env = append(cmd.Env, "VERSION="+version, "BUILD_KIND="+kind)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("make %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}

package tq

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	tqconfig "github.com/version-1/tasq/internal/config"
)

const (
	updateReleaseRepo = "version-1/tasq"
	updateInstallName = "tq"
)

type updateRunner interface {
	currentVersion(context.Context) (string, error)
	targetVersion(context.Context, string) (string, error)
	confirm(context.Context) (bool, error)
	stopServices(context.Context) error
	install(context.Context, string) error
	installedVersion(context.Context) (string, error)
	migrate(context.Context) error
	startServices(context.Context) error
}

var newUpdateRunner = func(a app, cfg config) updateRunner {
	cfg.output = "text"
	return defaultUpdateRunner{app: a, cfg: cfg}
}

func (a app) update(ctx context.Context, args []string, cfg config) error {
	if len(args) > 0 && (args[0] == "help" || args[0] == "-help" || args[0] == "--help") {
		printUpdateHelp(a.stdout)
		return nil
	}
	fs := newFlagSet("update")
	yes := fs.Bool("y", false, "update without confirmation")
	tag := fs.String("tag", "", "release tag to install")
	if err := fs.Parse(args); err != nil {
		return usageError(err.Error())
	}
	if fs.NArg() != 0 {
		return usageError("update does not accept positional arguments")
	}
	if os.Getenv(tqconfig.EnvManagedRun) == "1" {
		return errors.New("tq update is unavailable inside an orchestrator-managed run; run it from a user shell")
	}
	profile, err := tqconfig.DefaultHomeProfile()
	if err != nil {
		return err
	}
	if err := updateProfileAllowed(profile); err != nil {
		return err
	}

	runner := newUpdateRunner(a, cfg)
	current, err := runner.currentVersion(ctx)
	if err != nil {
		return fmt.Errorf("read current version: %w", err)
	}
	target, err := runner.targetVersion(ctx, *tag)
	if err != nil {
		return fmt.Errorf("resolve target version: %w", err)
	}
	fmt.Fprintf(a.stdout, "Current version: %s\n", current)
	fmt.Fprintf(a.stdout, "Target version:  %s\n", target)

	if !*yes {
		ok, err := runner.confirm(ctx)
		if err != nil {
			return err
		}
		if !ok {
			fmt.Fprintln(a.stdout, "Update cancelled")
			return nil
		}
	}

	if err := runUpdateStep(a.stdout, "Stopping services", func() error {
		return runner.stopServices(ctx)
	}); err != nil {
		return err
	}
	if err := runUpdateStep(a.stdout, "Installing tq", func() error {
		return runner.install(ctx, target)
	}); err != nil {
		return err
	}
	fmt.Fprintln(a.stdout, "Checking installed version...")
	installed, err := runner.installedVersion(ctx)
	if err != nil {
		return fmt.Errorf("check installed version: %w", err)
	}
	fmt.Fprintf(a.stdout, "Installed version: %s\n", installed)
	if err := runUpdateStep(a.stdout, "Migrating databases", func() error {
		return runner.migrate(ctx)
	}); err != nil {
		return err
	}
	if err := runUpdateStep(a.stdout, "Starting services", func() error {
		return runner.startServices(ctx)
	}); err != nil {
		return err
	}
	fmt.Fprintln(a.stdout, "tq update complete")
	return nil
}

func updateProfileAllowed(profile string) error {
	if profile == "" {
		return nil
	}
	return fmt.Errorf("tq update is unavailable for build profile %q; install a matching profile build instead", profile)
}

func runUpdateStep(w io.Writer, label string, fn func() error) error {
	fmt.Fprintf(w, "%s...\n", label)
	if err := fn(); err != nil {
		return fmt.Errorf("%s failed: %w", strings.ToLower(label), err)
	}
	return nil
}

type defaultUpdateRunner struct {
	app app
	cfg config
}

func (r defaultUpdateRunner) currentVersion(context.Context) (string, error) {
	version, commit := versionInfo()
	return fmt.Sprintf("tq %s (commit: %s)", version, commit), nil
}

func (r defaultUpdateRunner) targetVersion(ctx context.Context, tag string) (string, error) {
	if strings.TrimSpace(tag) != "" {
		return releaseTag(ctx, strings.TrimSpace(tag))
	}
	return latestFormalReleaseTag(ctx)
}

func (r defaultUpdateRunner) confirm(context.Context) (bool, error) {
	fmt.Fprint(r.app.stdout, "Updating tq will stop and restart local services. Continue? [y/N] ")
	line, err := bufio.NewReader(r.app.stdin).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, fmt.Errorf("read confirmation: %w", err)
	}
	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes", nil
}

func (r defaultUpdateRunner) stopServices(ctx context.Context) error {
	return r.app.serviceStop(ctx, nil, r.cfg)
}

func (r defaultUpdateRunner) install(ctx context.Context, tag string) error {
	return installRelease(ctx, tag, r.app.stdout)
}

func (r defaultUpdateRunner) installedVersion(ctx context.Context) (string, error) {
	path, err := installedTQPath()
	if err != nil {
		return "", err
	}
	output, err := exec.CommandContext(ctx, path, "version").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s version: %w: %s", path, err, strings.TrimSpace(string(output)))
	}
	return strings.TrimSpace(string(output)), nil
}

func (r defaultUpdateRunner) migrate(ctx context.Context) error {
	return r.app.migrateUp(ctx, r.cfg)
}

func (r defaultUpdateRunner) startServices(ctx context.Context) error {
	return r.app.serviceStart(ctx, nil, r.cfg)
}

func newGHCommand(ctx context.Context, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, "gh", args...)
	// Release updates must fail instead of waiting for an interactive prompt.
	cmd.Env = append(os.Environ(), "GH_PROMPT_DISABLED=1")
	return cmd
}

func latestFormalReleaseTag(ctx context.Context) (string, error) {
	output, err := newGHCommand(ctx, "release", "list",
		"--repo", updateReleaseRepo,
		"--exclude-drafts",
		"--exclude-pre-releases",
		"--limit", "1",
		"--json", "tagName",
		"--jq", ".[0].tagName // \"\"",
	).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("gh release list: %w: %s", err, strings.TrimSpace(string(output)))
	}
	tag := strings.TrimSpace(string(output))
	if tag == "" {
		return "", fmt.Errorf("no release found for %s", updateReleaseRepo)
	}
	return tag, nil
}

func releaseTag(ctx context.Context, tag string) (string, error) {
	output, err := newGHCommand(ctx, "release", "view", tag,
		"--repo", updateReleaseRepo,
		"--json", "tagName",
		"--jq", ".tagName",
	).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("gh release view %s: %w: %s", tag, err, strings.TrimSpace(string(output)))
	}
	resolved := strings.TrimSpace(string(output))
	if resolved == "" {
		return "", fmt.Errorf("release %s not found for %s", tag, updateReleaseRepo)
	}
	return resolved, nil
}

func installRelease(ctx context.Context, tag string, w io.Writer) error {
	platform, err := releasePlatform()
	if err != nil {
		return err
	}
	tmpDir, err := os.MkdirTemp("", "tq-update-*")
	if err != nil {
		return fmt.Errorf("create temporary directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	assetPattern := "tasq_*_" + platform + ".tar.gz"
	output, err := newGHCommand(ctx, "release", "download", tag,
		"--repo", updateReleaseRepo,
		"--pattern", assetPattern,
		"--pattern", "checksums.txt",
		"--dir", tmpDir,
		"--clobber",
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("gh release download %s: %w: %s", tag, err, strings.TrimSpace(string(output)))
	}

	assetPath, err := singleGlob(filepath.Join(tmpDir, assetPattern))
	if err != nil {
		return err
	}
	if err := verifyDownloadedAsset(assetPath, filepath.Join(tmpDir, "checksums.txt")); err != nil {
		return err
	}
	extractDir := filepath.Join(tmpDir, "extracted")
	if err := extractReleaseArchive(assetPath, extractDir); err != nil {
		return err
	}
	if err := installExtractedExecutables(extractDir); err != nil {
		return err
	}
	path, err := installedTQPath()
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "Installed tq from %s to %s\n", tag, path)
	return nil
}

func releasePlatform() (string, error) {
	var osName string
	switch runtime.GOOS {
	case "darwin", "linux":
		osName = runtime.GOOS
	default:
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
	var arch string
	switch runtime.GOARCH {
	case "amd64", "arm64":
		arch = runtime.GOARCH
	default:
		return "", fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
	}
	return osName + "_" + arch, nil
}

func singleGlob(pattern string) (string, error) {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("asset not found for pattern: %s", filepath.Base(pattern))
	}
	if len(matches) > 1 {
		return "", fmt.Errorf("multiple assets found for pattern: %s", filepath.Base(pattern))
	}
	return matches[0], nil
}

func verifyDownloadedAsset(assetPath string, checksumsPath string) error {
	checksums, err := os.ReadFile(checksumsPath)
	if err != nil {
		return fmt.Errorf("read checksums.txt: %w", err)
	}
	expected := checksumForAsset(string(checksums), filepath.Base(assetPath))
	if expected == "" {
		return fmt.Errorf("checksum not found for %s", filepath.Base(assetPath))
	}
	actual, err := sha256File(assetPath)
	if err != nil {
		return err
	}
	if expected != actual {
		return fmt.Errorf("checksum mismatch for %s", filepath.Base(assetPath))
	}
	return nil
}

func checksumForAsset(checksums string, assetName string) string {
	for _, line := range strings.Split(checksums, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[1] == assetName {
			return fields[0]
		}
	}
	return ""
}

func sha256File(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("hash %s: %w", path, err)
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func extractReleaseArchive(archivePath string, extractDir string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("open archive: %w", err)
	}
	defer file.Close()
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("open gzip archive: %w", err)
	}
	defer gzipReader.Close()

	if err := os.MkdirAll(extractDir, 0o755); err != nil {
		return fmt.Errorf("create extract dir: %w", err)
	}
	required := map[string]bool{
		"tq":            false,
		"issue-tracker": false,
		"orchestrator":  false,
		"web":           false,
	}
	reader := tar.NewReader(gzipReader)
	for {
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("read archive: %w", err)
		}
		name := filepath.Base(header.Name)
		if _, ok := required[name]; !ok || header.Typeflag != tar.TypeReg {
			continue
		}
		path := filepath.Join(extractDir, name)
		out, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
		if err != nil {
			return fmt.Errorf("create %s: %w", name, err)
		}
		if _, err := io.Copy(out, reader); err != nil {
			_ = out.Close()
			return fmt.Errorf("extract %s: %w", name, err)
		}
		if err := out.Close(); err != nil {
			return fmt.Errorf("close %s: %w", name, err)
		}
		required[name] = true
	}
	for name, found := range required {
		if !found {
			return fmt.Errorf("archive does not contain %s", name)
		}
	}
	return nil
}

func installExtractedExecutables(extractDir string) error {
	installDir, err := tqInstallDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(installDir, 0o755); err != nil {
		return fmt.Errorf("create install dir: %w", err)
	}
	home, err := tqconfig.Home()
	if err != nil {
		return err
	}
	serviceDir := serviceInstallDir(home)
	if err := os.MkdirAll(serviceDir, 0o755); err != nil {
		return fmt.Errorf("create service install dir: %w", err)
	}
	for _, executable := range []string{"issue-tracker", "orchestrator", "web"} {
		if err := installExecutable(filepath.Join(extractDir, executable), filepath.Join(serviceDir, executable)); err != nil {
			return err
		}
	}
	if err := installExecutable(filepath.Join(extractDir, "tq"), filepath.Join(installDir, updateInstallName)); err != nil {
		return err
	}
	return nil
}

func serviceInstallDir(home string) string {
	return filepath.Join(tqconfig.SystemDir(home), "bin")
}

func installExecutable(source string, destination string) error {
	input, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("open %s: %w", source, err)
	}
	defer input.Close()

	tmp := fmt.Sprintf("%s.tmp-%d", destination, os.Getpid())
	output, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return fmt.Errorf("create %s: %w", tmp, err)
	}
	if _, err := io.Copy(output, input); err != nil {
		_ = output.Close()
		_ = os.Remove(tmp)
		return fmt.Errorf("copy %s: %w", destination, err)
	}
	if err := output.Close(); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("close %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, destination); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("install %s: %w", destination, err)
	}
	return nil
}

func tqInstallDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home: %w", err)
	}
	return filepath.Join(home, ".local", "bin"), nil
}

func installedTQPath() (string, error) {
	dir, err := tqInstallDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, updateInstallName), nil
}

func printUpdateHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage: tq update [-y] [--tag TAG]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Flags:")
	fmt.Fprintln(w, "  -y         Update without confirmation")
	fmt.Fprintln(w, "  --tag TAG  Install a specific release or prerelease tag")
}

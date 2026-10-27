// Package release prepares local artifacts. It never creates tags or publishes.
package release

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

type Runner func(context.Context, string, []string, []string) ([]byte, error)
type Config struct {
	Root, Tag, Output, SourceArchive string
	Run                              Runner
}

var tagPattern = regexp.MustCompile(`^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)

func Command(ctx context.Context, name string, args, env []string) ([]byte, error) {
	c := exec.CommandContext(ctx, name, args...)
	c.Env = append(os.Environ(), env...)
	data, err := c.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s failed: %w\n%s", name, err, data)
	}
	return data, nil
}

func Prepare(ctx context.Context, c Config) error {
	if !tagPattern.MatchString(c.Tag) {
		return errors.New("tag must be an exact vMAJOR.MINOR.PATCH release (no dev or prerelease)")
	}
	if c.Run == nil {
		c.Run = Command
	}
	root, e := filepath.Abs(c.Root)
	if e != nil {
		return e
	}
	output, e := filepath.Abs(c.Output)
	if e != nil {
		return e
	}
	if c.Output == "" {
		return errors.New("an explicit new output directory is required")
	}
	if _, e = os.Lstat(output); !os.IsNotExist(e) {
		return errors.New("output directory must not already exist")
	}
	git := func(args ...string) ([]byte, error) {
		return c.Run(ctx, "git", append([]string{"-C", root}, args...), nil)
	}
	status, e := git("status", "--porcelain", "--untracked-files=all")
	if e != nil {
		return e
	}
	if len(bytes.TrimSpace(status)) > 0 {
		return errors.New("release requires a clean checkout, including untracked files")
	}
	head, e := git("rev-parse", "HEAD")
	if e != nil {
		return e
	}
	tagged, e := git("rev-parse", "--verify", "refs/tags/"+c.Tag+"^{commit}")
	if e != nil {
		return e
	}
	if !bytes.Equal(bytes.TrimSpace(head), bytes.TrimSpace(tagged)) {
		return errors.New("requested tag does not point to HEAD")
	}
	if c.SourceArchive == "" {
		return errors.New("provide the exact downloaded GitHub tag source archive with --source-archive")
	}
	files, e := git("ls-files", "-z")
	if e != nil {
		return e
	}
	source, e := os.ReadFile(c.SourceArchive)
	if e != nil {
		return e
	}
	if e = VerifySource(source, root, c.Tag, strings.Split(strings.TrimSuffix(string(files), "\x00"), "\x00")); e != nil {
		return e
	}
	compiler, e := c.Run(ctx, "go", []string{"version"}, nil)
	if e != nil {
		return e
	}
	// Run only maintainer checks and cross-builds, never a package manager.
	for _, args := range [][]string{{"-C", root, "test", "./..."}, {"-C", root, "vet", "./..."}} {
		if _, e = c.Run(ctx, "go", args, nil); e != nil {
			return e
		}
	}
	format, e := c.Run(ctx, "gofmt", []string{"-l", filepath.Join(root, "cmd"), filepath.Join(root, "internal")}, nil)
	if e != nil {
		return e
	}
	if len(bytes.TrimSpace(format)) > 0 {
		return errors.New("Go source is not formatted")
	}
	stage, e := os.MkdirTemp(filepath.Dir(output), ".upmacx-release-")
	if e != nil {
		return e
	}
	defer os.RemoveAll(stage)
	readme, e := os.ReadFile(filepath.Join(root, "README.md"))
	if e != nil {
		return e
	}
	license, e := os.ReadFile(filepath.Join(root, "LICENSE"))
	if e != nil {
		return e
	}
	var sums strings.Builder
	for _, arch := range []string{"arm64", "amd64"} {
		binary := filepath.Join(stage, "upmacx-"+arch)
		args := []string{"-C", root, "build", "-trimpath", "-ldflags=-s -w -X main.version=" + c.Tag, "-o", binary, "./cmd/upmacx"}
		if _, e = c.Run(ctx, "go", args, []string{"CGO_ENABLED=0", "GOOS=darwin", "GOARCH=" + arch}); e != nil {
			return e
		}
		executable, e := os.ReadFile(binary)
		if e != nil {
			return e
		}
		name := "upmacx_" + c.Tag + "_darwin_" + arch + ".tar.gz"
		archive, e := Archive(executable, readme, license)
		if e != nil {
			return e
		}
		if e = os.WriteFile(filepath.Join(stage, name), archive, 0644); e != nil {
			return e
		}
		fmt.Fprintf(&sums, "%x  %s\n", sha256.Sum256(archive), name)
		if e = os.Remove(binary); e != nil {
			return e
		}
	}
	if e = os.WriteFile(filepath.Join(stage, "checksums.txt"), []byte(sums.String()), 0644); e != nil {
		return e
	}
	formula := Formula(c.Tag, fmt.Sprintf("%x", sha256.Sum256(source)))
	if e = os.Mkdir(filepath.Join(stage, "Formula"), 0755); e != nil {
		return e
	}
	if e = os.WriteFile(filepath.Join(stage, "Formula", "upmacx.rb"), []byte(formula), 0644); e != nil {
		return e
	}
	metadata := fmt.Sprintf("Tag: %s\nCommit: %s\nCompiler: %s\nBuild: CGO_ENABLED=0 GOOS=darwin GOARCH=arm64,amd64 -trimpath -ldflags=-s -w -X main.version=%s\nSource SHA-256: %x\nNative provider/formula validation and publication are separate maintainer gates.\n", c.Tag, bytes.TrimSpace(head), bytes.TrimSpace(compiler), c.Tag, sha256.Sum256(source))
	if e = os.WriteFile(filepath.Join(stage, "build-info.txt"), []byte(metadata), 0644); e != nil {
		return e
	}
	// Verify final files from disk before exposing the complete output directory.
	for _, line := range strings.Split(strings.TrimSpace(sums.String()), "\n") {
		fields := strings.Fields(line)
		data, err := os.ReadFile(filepath.Join(stage, fields[1]))
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		if hex.EncodeToString(sum[:]) != fields[0] {
			return errors.New("archive checksum verification failed")
		}
	}
	return os.Rename(stage, output)
}

func Archive(binary, readme, license []byte) ([]byte, error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	tw := tar.NewWriter(gz)
	for _, f := range []struct {
		name string
		mode int64
		data []byte
	}{{"upmacx", 0755, binary}, {"README", 0644, readme}, {"LICENSE", 0644, license}} {
		if err := tw.WriteHeader(&tar.Header{Name: f.name, Mode: f.mode, Size: int64(len(f.data)), Typeflag: tar.TypeReg}); err != nil {
			return nil, err
		}
		if _, err := tw.Write(f.data); err != nil {
			return nil, err
		}
	}
	if e := tw.Close(); e != nil {
		return nil, e
	}
	if e := gz.Close(); e != nil {
		return nil, e
	}
	return b.Bytes(), nil
}

// VerifySource compares every tracked file with the downloaded tag archive,
// rejecting links, traversal, duplicates, missing files and extra file content.
func VerifySource(source []byte, root, tag string, tracked []string) error {
	expected := map[string]bool{}
	for _, name := range tracked {
		if name == "" {
			return errors.New("empty tracked source list")
		}
		expected[name] = true
	}
	gz, e := gzip.NewReader(bytes.NewReader(source))
	if e != nil {
		return e
	}
	defer gz.Close()
	tr := tar.NewReader(io.LimitReader(gz, 256<<20))
	prefix := "upmacx-" + strings.TrimPrefix(tag, "v") + "/"
	seen := map[string]bool{}
	for {
		h, e := tr.Next()
		if e == io.EOF {
			break
		}
		if e != nil {
			return e
		}
		if h.Typeflag == tar.TypeXGlobalHeader {
			continue
		}
		if !strings.HasPrefix(h.Name, prefix) {
			return fmt.Errorf("unexpected source archive path: %s", h.Name)
		}
		name := strings.TrimPrefix(h.Name, prefix)
		if name == "" && h.Typeflag == tar.TypeDir {
			continue
		}
		if path.IsAbs(name) || path.Clean(name) != strings.TrimSuffix(name, "/") || strings.HasPrefix(name, "../") {
			return errors.New("unsafe source archive path")
		}
		if h.Typeflag == tar.TypeDir {
			continue
		}
		if h.Typeflag != tar.TypeReg || !expected[name] || seen[name] || h.Size > 16<<20 {
			return fmt.Errorf("unexpected source archive entry: %s", h.Name)
		}
		info, e := os.Lstat(filepath.Join(root, filepath.FromSlash(name)))
		if e != nil {
			return e
		}
		if !info.Mode().IsRegular() {
			return errors.New("source must contain regular tracked files")
		}
		local, e := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
		if e != nil {
			return e
		}
		data, e := io.ReadAll(tr)
		if e != nil {
			return e
		}
		if !bytes.Equal(local, data) {
			return fmt.Errorf("source archive differs from tagged checkout: %s", name)
		}
		seen[name] = true
	}
	// tar EOF can precede the gzip footer. Drain it to verify the checksum.
	if _, e = io.Copy(io.Discard, gz); e != nil {
		return e
	}
	if len(seen) != len(expected) {
		return errors.New("source archive is missing tracked files")
	}
	return nil
}
func Formula(tag, checksum string) string {
	return fmt.Sprintf(`class Upmacx < Formula
  desc "Your macOS update assistant"
  homepage "https://github.com/htooaunglynn/upmacx"
  url "https://github.com/htooaunglynn/upmacx/archive/refs/tags/%s.tar.gz"
  sha256 "%s"
  license "MIT"

  depends_on "go" => :build
  depends_on :macos

  def install
    system "go", "build", "-trimpath",
           "-ldflags=-s -w -X main.version=v#{version}",
           "-o", bin/"upmacx", "./cmd/upmacx"
  end

  test do
    assert_equal "upmacx v#{version}\n", shell_output("#{bin}/upmacx --version")
    assert_match "Usage: upmacx", shell_output("#{bin}/upmacx --help")
  end
end
`, tag, checksum)
}

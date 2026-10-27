package release

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func sourceArchive(t *testing.T, entries []struct {
	name, value string
	kind        byte
}) []byte {
	t.Helper()
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	tw := tar.NewWriter(gz)
	for _, entry := range entries {
		h := &tar.Header{Name: entry.name, Mode: 0644, Size: int64(len(entry.value)), Typeflag: entry.kind}
		if entry.kind == tar.TypeSymlink {
			h.Linkname = entry.value
			h.Size = 0
		}
		if err := tw.WriteHeader(h); err != nil {
			t.Fatal(err)
		}
		if entry.kind != tar.TypeSymlink {
			if _, err := tw.Write([]byte(entry.value)); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}
func TestSourceValidation(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example"), 0644); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name    string
		entries []struct {
			name, value string
			kind        byte
		}
		valid bool
	}{
		{"valid", []struct {
			name, value string
			kind        byte
		}{{"upmacx-0.1.0/go.mod", "module example", tar.TypeReg}}, true},
		{"different content", []struct {
			name, value string
			kind        byte
		}{{"upmacx-0.1.0/go.mod", "module wrong", tar.TypeReg}}, false},
		{"wrong tag", []struct {
			name, value string
			kind        byte
		}{{"upmacx-0.2.0/go.mod", "module example", tar.TypeReg}}, false},
		{"traversal", []struct {
			name, value string
			kind        byte
		}{{"upmacx-0.1.0/../go.mod", "module example", tar.TypeReg}}, false},
		{"symlink", []struct {
			name, value string
			kind        byte
		}{{"upmacx-0.1.0/go.mod", "/tmp/elsewhere", tar.TypeSymlink}}, false},
		{"duplicate", []struct {
			name, value string
			kind        byte
		}{{"upmacx-0.1.0/go.mod", "module example", tar.TypeReg}, {"upmacx-0.1.0/go.mod", "module example", tar.TypeReg}}, false},
		{"missing", nil, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := VerifySource(sourceArchive(t, tc.entries), root, "v0.1.0", []string{"go.mod"})
			if (e == nil) != tc.valid {
				t.Fatal(e)
			}
		})
	}
}
func TestArchive(t *testing.T) {
	data, e := Archive([]byte("executable"), []byte("docs"), []byte("MIT"))
	if e != nil {
		t.Fatal(e)
	}
	again, _ := Archive([]byte("executable"), []byte("docs"), []byte("MIT"))
	if !bytes.Equal(data, again) {
		t.Fatal("unstable archive bytes for identical input")
	}
	gz, e := gzip.NewReader(bytes.NewReader(data))
	if e != nil {
		t.Fatal(e)
	}
	defer gz.Close()
	r := tar.NewReader(gz)
	want := []string{"upmacx", "README", "LICENSE"}
	for i, name := range want {
		h, e := r.Next()
		if e != nil || h.Name != name {
			t.Fatal(h, e)
		}
		content, e := io.ReadAll(r)
		if e != nil {
			t.Fatal(e)
		}
		if i == 0 && (h.Mode != 0755 || string(content) != "executable") {
			t.Fatal("invalid executable entry")
		}
	}
	if _, e = r.Next(); e != io.EOF {
		t.Fatal(e)
	}
}
func TestPrepare(t *testing.T) {
	root := t.TempDir()
	tracked := []string{"README.md", "LICENSE", "go.mod"}
	var entries []struct {
		name, value string
		kind        byte
	}
	for _, name := range tracked {
		data := "content of " + name
		if e := os.WriteFile(filepath.Join(root, name), []byte(data), 0644); e != nil {
			t.Fatal(e)
		}
		entries = append(entries, struct {
			name, value string
			kind        byte
		}{"upmacx-0.1.0/" + name, data, tar.TypeReg})
	}
	source := sourceArchive(t, entries)
	sourcePath := filepath.Join(t.TempDir(), "source.tar.gz")
	if e := os.WriteFile(sourcePath, source, 0644); e != nil {
		t.Fatal(e)
	}
	for _, scenario := range []string{"success", "dirty", "wrong tag", "build failure", "test failure", "unformatted", "existing output", "missing source", "bad tag"} {
		t.Run(scenario, func(t *testing.T) {
			out := filepath.Join(t.TempDir(), "artifacts")
			c := Config{Root: root, Tag: "v0.1.0", Output: out, SourceArchive: sourcePath}
			var builds []string
			if scenario == "existing output" {
				if e := os.Mkdir(out, 0755); e != nil {
					t.Fatal(e)
				}
			}
			if scenario == "missing source" {
				c.SourceArchive = ""
			}
			if scenario == "bad tag" {
				c.Tag = "dev"
			}
			c.Run = func(_ context.Context, name string, args, env []string) ([]byte, error) {
				switch name {
				case "git":
					switch args[2] {
					case "status":
						if scenario == "dirty" {
							return []byte("?? untracked"), nil
						}
						return nil, nil
					case "rev-parse":
						if len(args) > 3 && args[3] == "--verify" && scenario == "wrong tag" {
							return []byte("other"), nil
						}
						return []byte("commit"), nil
					case "ls-files":
						return []byte(strings.Join(tracked, "\x00") + "\x00"), nil
					}
				case "gofmt":
					if scenario == "unformatted" {
						return []byte("file.go"), nil
					}
					return nil, nil
				case "go":
					if args[0] == "version" {
						return []byte("go version test"), nil
					}
					if args[2] == "test" && scenario == "test failure" {
						return nil, errors.New("test failed")
					}
					if args[2] == "build" {
						if scenario == "build failure" {
							return nil, errors.New("build failed")
						}
						if !reflect.DeepEqual(env, []string{"CGO_ENABLED=0", "GOOS=darwin", "GOARCH=" + []string{"arm64", "amd64"}[len(builds)]}) {
							t.Fatal(env)
						}
						if !strings.Contains(strings.Join(args, " "), "-X main.version=v0.1.0") {
							t.Fatal(args)
						}
						var binary string
						for i, a := range args {
							if a == "-o" {
								binary = args[i+1]
							}
						}
						builds = append(builds, binary)
						return nil, os.WriteFile(binary, []byte("fake cross-built executable"), 0755)
					}
					return nil, nil
				}
				return nil, fmt.Errorf("unexpected command %s %v", name, args)
			}
			e := Prepare(context.Background(), c)
			if scenario != "success" {
				if e == nil {
					t.Fatal("expected rejection")
				}
				if scenario != "existing output" {
					if _, e = os.Stat(out); !os.IsNotExist(e) {
						t.Fatal("left partial output")
					}
				}
				return
			}
			if e != nil {
				t.Fatal(e)
			}
			if len(builds) != 2 {
				t.Fatal(builds)
			}
			sums, e := os.ReadFile(filepath.Join(out, "checksums.txt"))
			if e != nil {
				t.Fatal(e)
			}
			for _, line := range strings.Split(strings.TrimSpace(string(sums)), "\n") {
				fields := strings.Fields(line)
				data, e := os.ReadFile(filepath.Join(out, fields[1]))
				if e != nil {
					t.Fatal(e)
				}
				if fields[0] != fmt.Sprintf("%x", sha256.Sum256(data)) {
					t.Fatal("wrong checksum")
				}
			}
			formula, e := os.ReadFile(filepath.Join(out, "Formula", "upmacx.rb"))
			if e != nil {
				t.Fatal(e)
			}
			for _, s := range []string{fmt.Sprintf("%x", sha256.Sum256(source)), "refs/tags/v0.1.0.tar.gz", "depends_on :macos", "--version", "--help"} {
				if !strings.Contains(string(formula), s) {
					t.Fatal("missing formula requirement", s)
				}
			}
		})
	}
}

func TestSourceRejectsCorruptGzipFooter(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module test"), 0644); err != nil {
		t.Fatal(err)
	}
	data := sourceArchive(t, []struct {
		name, value string
		kind        byte
	}{{"upmacx-0.1.0/go.mod", "module test", tar.TypeReg}})
	data[len(data)-8] ^= 0xff
	if err := VerifySource(data, root, "v0.1.0", []string{"go.mod"}); err == nil {
		t.Fatal("accepted corrupt gzip checksum")
	}
}

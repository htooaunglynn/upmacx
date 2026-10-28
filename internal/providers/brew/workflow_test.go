package brew

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/htooaunglynn/upmacx/internal/cli"
	"github.com/htooaunglynn/upmacx/internal/process"
	"github.com/htooaunglynn/upmacx/internal/updater"
)

// TestMain turns this test executable into the only package manager on fake PATH.
func TestMain(m *testing.M) {
	if os.Getenv("UPMACX_FAKE_BREW") == "1" {
		fakeBrew()
		os.Exit(0)
	}
	os.Exit(m.Run())
}
func fakeBrew() {
	f, e := os.OpenFile(os.Getenv("UPMACX_LOG"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if e != nil {
		os.Exit(90)
	}
	dir, _ := os.Getwd()
	_ = json.NewEncoder(f).Encode(struct {
		Args                                  []string
		Dir, Clean, Auto, Greedy, GreedyCasks string
	}{os.Args[1:], dir, os.Getenv("HOMEBREW_NO_INSTALL_CLEANUP"), os.Getenv("HOMEBREW_NO_AUTO_UPDATE"), os.Getenv("HOMEBREW_UPGRADE_GREEDY"), os.Getenv("HOMEBREW_UPGRADE_GREEDY_CASKS")})
	_ = f.Close()
	command := strings.Join(os.Args[1:], " ")
	if command == os.Getenv("UPMACX_FAIL") {
		fmt.Fprintln(os.Stderr, "simulated permission denied")
		os.Exit(1)
	}
	switch command {
	case "--version":
		fmt.Println("Homebrew 7.0.10")
	case "help outdated":
		fmt.Println("--json v2 --greedy")
	case "help list":
		fmt.Println("--pinned --cask")
	case "help upgrade":
		fmt.Println("--formula --cask")
	case "help cleanup":
		fmt.Println("cleanup")
	case "update", "list --pinned", "upgrade --formula tool", "upgrade --cask app", "cleanup":
	case "outdated --json=v2":
		fmt.Print(os.Getenv("UPMACX_DATA"))
	default:
		fmt.Fprintln(os.Stderr, "unexpected fake executable invocation", command)
		os.Exit(91)
	}
}
func TestFakeBrewWorkflow(t *testing.T) {
	for _, tc := range []struct {
		name               string
		args               []string
		answer, fail, data string
		terminal           bool
		code               int
		mutations          string
	}{
		{name: "check", args: []string{"--check"}},
		{name: "decline", answer: "no\n", terminal: true},
		{name: "no tty", code: 2},
		{name: "approve", args: []string{"--yes"}, mutations: "upgrade --formula tool|upgrade --cask app"},
		{name: "cleanup", args: []string{"--yes", "--clean"}, mutations: "upgrade --formula tool|upgrade --cask app|cleanup"},
		{name: "partial permission failure", args: []string{"--yes", "--clean"}, fail: "upgrade --cask app", code: 1, mutations: "upgrade --formula tool|upgrade --cask app"},
		{name: "refresh failure", args: []string{"--yes", "--clean"}, fail: "update", code: 1},
		{name: "cleanup failure", args: []string{"--yes", "--clean"}, fail: "cleanup", code: 1, mutations: "upgrade --formula tool|upgrade --cask app|cleanup"},
		{name: "empty", data: `{"formulae":[],"casks":[]}`},
		{name: "malformed", data: `{"formulae":[`, code: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			exe, e := os.Executable()
			if e != nil {
				t.Fatal(e)
			}
			if e = os.Symlink(exe, filepath.Join(dir, "brew")); e != nil {
				t.Fatal(e)
			}
			log := filepath.Join(dir, "commands.jsonl")
			data := tc.data
			if data == "" {
				data = `{"formulae":[{"name":"tool","installed_versions":["1"],"current_version":"2","pinned":false}],"casks":[{"name":"app","installed_versions":["1"],"current_version":"2","pinned":false}]}`
			}
			for k, v := range map[string]string{"PATH": dir, "HOME": dir, "HOMEBREW_PREFIX": dir, "HOMEBREW_UPGRADE_GREEDY": "1", "HOMEBREW_UPGRADE_GREEDY_CASKS": "app", "UPMACX_FAKE_BREW": "1", "UPMACX_LOG": log, "UPMACX_DATA": data, "UPMACX_FAIL": tc.fail, "GORACE": "atexit_sleep_ms=0"} {
				t.Setenv(k, v)
			}
			var out, errout bytes.Buffer
			runner := &process.Executor{Dir: dir, In: strings.NewReader(""), Out: &out, Err: &errout}
			app := cli.App{Arch: "arm64", OS: "darwin", UID: 501, In: strings.NewReader(tc.answer), Out: &out, Err: &errout, Terminal: func() bool { return tc.terminal }, Providers: []updater.Provider{New(runner)}}
			if code := app.Run(context.Background(), tc.args); code != tc.code {
				t.Fatalf("exit %d want %d: %s", code, tc.code, errout.String())
			}
			file, e := os.Open(log)
			if e != nil {
				t.Fatal(e)
			}
			defer file.Close()
			d := json.NewDecoder(file)
			var mutations []string
			for d.More() {
				var entry struct {
					Args                                  []string
					Dir, Clean, Auto, Greedy, GreedyCasks string
				}
				if e = d.Decode(&entry); e != nil {
					t.Fatal(e)
				}
				if entry.Clean != "1" || entry.Auto != "1" || entry.Greedy != "0" || entry.GreedyCasks != "" {
					t.Fatal(entry)
				}
				a := strings.Join(entry.Args, " ")
				if strings.HasPrefix(a, "upgrade ") || a == "cleanup" {
					mutations = append(mutations, a)
				}
			}
			if strings.Join(mutations, "|") != tc.mutations {
				t.Fatalf("mutations: %v", mutations)
			}
			if tc.fail != "" && !strings.Contains(errout.String(), "permission denied") && tc.fail != "cleanup" && tc.fail != "upgrade --cask app" {
				t.Fatal("missing captured diagnostics", errout.String())
			}
		})
	}
}

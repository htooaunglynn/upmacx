package brew

import (
	"context"
	"errors"
	"github.com/htooaunglynn/upmacx/internal/process"
	"github.com/htooaunglynn/upmacx/internal/updater"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

type runner struct {
	commands []process.Command
	missing  bool
	fail     string
	data     []byte
	version  string
}

func (r *runner) LookPath(string) (string, error) {
	if r.missing {
		return "", exec.ErrNotFound
	}
	return "/fake/brew", nil
}
func (r *runner) Run(_ context.Context, c process.Command) (process.CommandResult, error) {
	r.commands = append(r.commands, c)
	s := strings.Join(c.Args, " ")
	if s == r.fail {
		return process.CommandResult{ExitCode: 1}, errors.New("fake command failure")
	}
	out := ""
	switch s {
	case "--version":
		out = "Homebrew 7.0.10"
		if r.version != "" {
			out = r.version
		}
	case "help outdated":
		out = "--json v2 --greedy"
	case "help list":
		out = "--pinned --cask"
	case "help upgrade":
		out = "--formula --cask"
	case "help cleanup":
		out = "cleanup"
	case "list --pinned":
		out = "held\n"
	case "outdated --json=v2":
		out = string(r.data)
	}
	return process.CommandResult{Stdout: []byte(out)}, nil
}
func fixture(t *testing.T) []byte {
	t.Helper()
	b, e := os.ReadFile("testdata/mixed.json")
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func TestWorkflow(t *testing.T) {
	r := &runner{data: fixture(t)}
	p := New(r)
	a, e := p.Detect(context.Background())
	if e != nil || !a.Available {
		t.Fatal(a, e)
	}
	plan, e := p.Discover(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	if len(plan.Candidates) != 2 || len(plan.Exclusions) != 1 {
		t.Fatal(plan)
	}
	o, e := p.Apply(context.Background(), plan)
	if e != nil || len(o.Completed) != 2 {
		t.Fatal(o, e)
	}
	if e = p.Cleanup(context.Background()); e != nil {
		t.Fatal(e)
	}
	var commands []string
	for _, c := range r.commands {
		commands = append(commands, strings.Join(c.Args, " "))
		if c.Path != "/fake/brew" || c.Env["HOMEBREW_NO_INSTALL_CLEANUP"] != "1" || c.Env["HOMEBREW_NO_AUTO_UPDATE"] != "1" || c.Env["HOMEBREW_UPGRADE_GREEDY"] != "0" || c.Env["HOMEBREW_UPGRADE_GREEDY_CASKS"] != "" {
			t.Fatal(c)
		}
	}
	want := []string{"--version", "help outdated", "help list", "help upgrade", "help cleanup", "update", "list --pinned", "outdated --json=v2", "upgrade --formula example/tap/tool", "upgrade --cask example-app", "cleanup"}
	if !reflect.DeepEqual(commands, want) {
		t.Fatal(commands)
	}
}
func TestMalformed(t *testing.T) {
	for _, s := range []string{"", `{}`, `{"formulae":null,"casks":[]}`, `{"formulae":[],"casks":[{"name":"-bad","installed_versions":["1"],"current_version":"2","pinned":false}]}`, `{"formulae":[{"name":"ok","installed_versions":["1"],"current_version":"2"}],"casks":[]}`, `{"formulae":[],"casks":[]} garbage`} {
		if _, e := Parse([]byte(s), nil); e == nil {
			t.Fatalf("accepted %s", s)
		}
	}
}
func TestEmptyAndKinds(t *testing.T) {
	for _, s := range []string{`{"formulae":[],"casks":[]}`, `{"formulae":[{"name":"tool","installed_versions":["1"],"current_version":"2","pinned":false}],"casks":[]}`, `{"formulae":[],"casks":[{"name":"app","installed_versions":["1"],"current_version":"2","pinned":false}]}`} {
		if _, e := Parse([]byte(s), nil); e != nil {
			t.Fatal(e)
		}
	}
}
func TestUnavailable(t *testing.T) {
	for _, r := range []*runner{{missing: true}, {version: "Homebrew unknown"}, {fail: "help outdated"}} {
		a, e := New(r).Detect(context.Background())
		if e != nil || a.Available {
			t.Fatal(a, e)
		}
	}
}
func TestFailures(t *testing.T) {
	for _, stage := range []string{"update", "list --pinned", "outdated --json=v2", "upgrade --cask example-app", "cleanup"} {
		r := &runner{data: fixture(t), fail: stage}
		p := New(r)
		_, _ = p.Detect(context.Background())
		plan, e := p.Discover(context.Background())
		if strings.HasPrefix(stage, "upgrade") {
			out, err := p.Apply(context.Background(), plan)
			if err == nil || len(out.Completed) != 1 {
				t.Fatal(out, err)
			}
		} else if stage == "cleanup" {
			if p.Cleanup(context.Background()) == nil {
				t.Fatal("expected failure")
			}
		} else if e == nil {
			t.Fatal("expected discovery failure")
		}
	}
}
func TestUnsafePlan(t *testing.T) {
	p := New(&runner{})
	for _, plan := range []updater.Plan{{Provider: "brew"}, {Provider: "brew", Candidates: []updater.Candidate{{ID: "--all", Kind: "formula"}}}} {
		if _, e := p.Apply(context.Background(), plan); e == nil {
			t.Fatal("accepted unsafe plan")
		}
	}
}

type failingProbe struct {
	runner
	failure error
	exit    int
}

func (r *failingProbe) Run(ctx context.Context, c process.Command) (process.CommandResult, error) {
	if strings.Join(c.Args, " ") == "help outdated" {
		return process.CommandResult{ExitCode: r.exit}, r.failure
	}
	return r.runner.Run(ctx, c)
}
func TestProbeRuntimeFailures(t *testing.T) {
	for _, e := range []error{context.Canceled, context.DeadlineExceeded, process.ErrOverflow, os.ErrPermission} {
		t.Run(e.Error(), func(t *testing.T) {
			r := &failingProbe{failure: e, exit: -1}
			a, err := New(r).Detect(context.Background())
			if a.Available || !errors.Is(err, e) {
				t.Fatalf("silenced runtime failure: %v %v", a, err)
			}
		})
	}
}

func TestVersionWithTapDetails(t *testing.T) {
	r := &runner{version: "Homebrew 7.0.10\nHomebrew/homebrew-core N/A\n"}
	a, err := New(r).Detect(context.Background())
	if err != nil || !a.Available || a.Version != "Homebrew 7.0.10" {
		t.Fatal(a, err)
	}
}

func TestNativeCaskSnapshots(t *testing.T) {
	data, err := os.ReadFile("testdata/homebrew-7.0.10-darwin-arm64-casks.json")
	if err != nil {
		t.Fatal(err)
	}
	plan, err := Parse(data, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Candidates) != 1 || plan.Candidates[0].ID != "fixture-cask" || plan.Candidates[0].Installed != "1.0" || plan.Candidates[0].Target != "2.0" {
		t.Fatal(plan)
	}
	if len(plan.Exclusions) != 1 || plan.Exclusions[0].Name != "fixture-cask-held" || plan.Exclusions[0].Reason != "pinned" {
		t.Fatal(plan.Exclusions)
	}
	data, err = os.ReadFile("testdata/homebrew-7.0.10-darwin-arm64-empty.json")
	if err != nil {
		t.Fatal(err)
	}
	plan, err = Parse(data, nil)
	if err != nil || len(plan.Candidates) != 0 || len(plan.Exclusions) != 0 {
		t.Fatal(plan, err)
	}
}

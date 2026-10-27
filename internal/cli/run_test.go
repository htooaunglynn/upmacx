package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"reflect"
	"strings"
	"syscall"
	"testing"

	"github.com/htooaunglynn/upmacx/internal/process"
	"github.com/htooaunglynn/upmacx/internal/updater"
)

type fake struct {
	id                                                      string
	log                                                     *[]string
	missing, detectFail, discoverFail, applyFail, cleanFail bool
	candidates                                              int
	cancel                                                  context.CancelCauseFunc
}

func (f *fake) ID() string { return f.id }
func (f *fake) Detect(context.Context) (updater.Availability, error) {
	*f.log = append(*f.log, f.id+":detect")
	if f.detectFail {
		return updater.Availability{}, errors.New("probe failed")
	}
	return updater.Availability{Available: !f.missing, Reason: "missing"}, nil
}
func (f *fake) Discover(context.Context) (updater.Plan, error) {
	*f.log = append(*f.log, f.id+":discover")
	if f.discoverFail {
		return updater.Plan{}, errors.New("discovery failed")
	}
	p := updater.Plan{Provider: f.id}
	for range f.candidates {
		p.Candidates = append(p.Candidates, updater.Candidate{ID: "one", Name: "one\x1b", Target: "2", Installed: "1"})
	}
	return p, nil
}
func (f *fake) Apply(context.Context, updater.Plan) (updater.Outcome, error) {
	*f.log = append(*f.log, f.id+":apply")
	o := updater.Outcome{Completed: []string{"one"}}
	if f.cancel != nil {
		f.cancel(process.SignalError{Signal: syscall.SIGINT})
		return o, context.Canceled
	}
	if f.applyFail {
		return o, errors.New("upgrade failed")
	}
	return o, nil
}
func (f *fake) Cleanup(context.Context) error {
	*f.log = append(*f.log, f.id+":cleanup")
	if f.cleanFail {
		return errors.New("cleanup failed")
	}
	return nil
}

type forbiddenRead struct{}

func (forbiddenRead) Read([]byte) (int, error) { panic("nonterminal input read") }
func TestParse(t *testing.T) {
	for _, args := range [][]string{{"--wat"}, {"positional"}, {"--providers", ""}, {"--providers", "brew,"}, {"--providers", "npm"}, {"--providers", "brew", "--providers", "brew"}, {"--check", "--clean"}, {"--check", "--system"}, {"--system"}, {"--no-clean"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			if _, e := Parse(args); e == nil {
				t.Fatal("accepted invalid arguments")
			}
		})
	}
	for _, args := range [][]string{{"--yes", "--check"}, {"--providers", "brew,brew"}, {"--help", "--version"}, {"--version", "--clean"}} {
		if _, e := Parse(args); e != nil {
			t.Fatal(e)
		}
	}
}
func TestRunMatrix(t *testing.T) {
	for _, tc := range []struct {
		name                                                            string
		args                                                            []string
		answer                                                          string
		terminal, missing, explicit, discoverFail, applyFail, cleanFail bool
		n, code                                                         int
		mutation                                                        []string
	}{
		{name: "check", args: []string{"--check"}, n: 1},
		{name: "yes", args: []string{"--yes"}, n: 1, mutation: []string{"brew:apply"}},
		{name: "consent", answer: " YeS \n", terminal: true, n: 1, mutation: []string{"brew:apply"}},
		{name: "no", answer: "no\n", terminal: true, n: 1},
		{name: "empty", answer: "\n", terminal: true, n: 1},
		{name: "eof", answer: "yes", terminal: true, n: 1},
		{name: "other", answer: "ok\n", terminal: true, n: 1},
		{name: "nonterminal", n: 1, code: 2},
		{name: "nothing"},
		{name: "cleanup empty", args: []string{"--yes", "--clean"}, mutation: []string{"brew:cleanup"}},
		{name: "cleanup decline", args: []string{"--clean"}, terminal: true, answer: "n\n"},
		{name: "cleanup success", args: []string{"--yes", "--clean"}, n: 1, mutation: []string{"brew:apply", "brew:cleanup"}},
		{name: "apply failure", args: []string{"--yes", "--clean"}, n: 1, applyFail: true, code: 1, mutation: []string{"brew:apply"}},
		{name: "cleanup failure", args: []string{"--yes", "--clean"}, n: 1, cleanFail: true, code: 1, mutation: []string{"brew:apply", "brew:cleanup"}},
		{name: "discovery failure", args: []string{"--yes", "--clean"}, discoverFail: true, code: 1},
		{name: "default missing", missing: true},
		{name: "explicit missing", args: []string{"--providers", "brew"}, missing: true, code: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var log []string
			var out, errout bytes.Buffer
			f := &fake{id: "brew", log: &log, candidates: tc.n, missing: tc.missing, discoverFail: tc.discoverFail, applyFail: tc.applyFail, cleanFail: tc.cleanFail}
			var in io.Reader = forbiddenRead{}
			if tc.terminal {
				in = strings.NewReader(tc.answer)
			}
			a := App{Version: "dev", OS: "darwin", UID: 501, In: in, Out: &out, Err: &errout, Terminal: func() bool { return tc.terminal }, Providers: []updater.Provider{f}}
			if code := a.Run(context.Background(), tc.args); code != tc.code {
				t.Fatalf("code %d want %d: %s", code, tc.code, errout.String())
			}
			var mutations []string
			for _, s := range log {
				if strings.HasSuffix(s, ":apply") || strings.HasSuffix(s, ":cleanup") {
					mutations = append(mutations, s)
				}
			}
			if !reflect.DeepEqual(mutations, tc.mutation) {
				t.Fatalf("mutations %v want %v", mutations, tc.mutation)
			}
			if strings.Contains(out.String(), "\x1b") {
				t.Fatal("unsanitized output")
			}
		})
	}
}
func TestInfoAndRefusal(t *testing.T) {
	for _, tc := range []struct {
		os   string
		uid  int
		args []string
		code int
	}{{"linux", 0, []string{"--help", "--version"}, 0}, {"linux", 0, []string{"--version"}, 0}, {"linux", 501, nil, 1}, {"darwin", 0, nil, 1}, {"darwin", 501, []string{"--unknown"}, 2}} {
		var log []string
		var b bytes.Buffer
		a := App{OS: tc.os, UID: tc.uid, Version: "v0.1.0", Out: &b, Err: &b, Providers: []updater.Provider{&fake{id: "brew", log: &log}}}
		if got := a.Run(context.Background(), tc.args); got != tc.code {
			t.Fatal(got)
		}
		if len(log) > 0 {
			t.Fatal(log)
		}
	}
}
func TestDiscoveryBeforeMutationAndFailure(t *testing.T) {
	var log []string
	var b bytes.Buffer
	a := App{OS: "darwin", UID: 501, Out: &b, Err: &b, Providers: []updater.Provider{&fake{id: "brew", log: &log, candidates: 1, applyFail: true}, &fake{id: "second", log: &log, candidates: 1}}}
	if a.Run(context.Background(), []string{"--yes"}) != 1 {
		t.Fatal("expected failure")
	}
	want := []string{"brew:detect", "brew:discover", "second:detect", "second:discover", "brew:apply", "second:apply"}
	if !reflect.DeepEqual(log, want) {
		t.Fatal(log)
	}
	if !strings.Contains(b.String(), "completed: one") {
		t.Fatal("lost partial result")
	}
}
func TestSignal(t *testing.T) {
	for _, sig := range []syscall.Signal{syscall.SIGINT, syscall.SIGTERM} {
		ctx, cancel := context.WithCancelCause(context.Background())
		cancel(process.SignalError{Signal: sig})
		var b bytes.Buffer
		var log []string
		a := App{OS: "darwin", UID: 501, Out: &b, Err: &b, Providers: []updater.Provider{&fake{id: "brew", log: &log}}}
		want := 130
		if sig == syscall.SIGTERM {
			want = 143
		}
		if got := a.Run(ctx, nil); got != want {
			t.Fatal(got)
		}
		if len(log) != 0 {
			t.Fatal(log)
		}
	}
}

func TestCancellationPreservesPriorFailure(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	var log []string
	var b bytes.Buffer
	a := App{OS: "darwin", UID: 501, Out: &b, Err: &b, Providers: []updater.Provider{
		&fake{id: "brew", log: &log, discoverFail: true},
		&fake{id: "second", log: &log, candidates: 1, cancel: cancel},
		&fake{id: "third", log: &log, candidates: 1},
	}}
	if got := a.Run(ctx, []string{"--yes"}); got != 130 {
		t.Fatal(got)
	}
	for _, want := range []string{"[brew] failed\n", "[second] interrupted (completed: one)", "[third] interrupted"} {
		if !strings.Contains(b.String(), want) {
			t.Fatalf("missing %q in %s", want, b.String())
		}
	}
	for _, entry := range log {
		if entry == "third:apply" {
			t.Fatal("scheduled mutation after cancellation")
		}
	}
}

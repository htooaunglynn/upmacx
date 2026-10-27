package updater

import (
	"context"
	"reflect"
	"syscall"
	"testing"

	"github.com/htooaunglynn/upmacx/internal/process"
)

type testProvider struct {
	id         string
	events     *[]string
	candidates []Candidate
	discover   func()
	apply      func()
	applied    []Candidate
}

func (p *testProvider) ID() string { return p.id }
func (p *testProvider) Detect(context.Context) (Availability, error) {
	*p.events = append(*p.events, p.id+":detect")
	return Availability{Available: true}, nil
}
func (p *testProvider) Discover(context.Context) (Plan, error) {
	*p.events = append(*p.events, p.id+":discover")
	if p.discover != nil {
		p.discover()
	}
	return Plan{Provider: p.id, Candidates: p.candidates}, nil
}
func (p *testProvider) Apply(_ context.Context, plan Plan) (Outcome, error) {
	*p.events = append(*p.events, p.id+":apply")
	p.applied = plan.Candidates
	if p.apply != nil {
		p.apply()
	}
	return Outcome{Completed: []string{plan.Candidates[0].ID}}, nil
}
func (p *testProvider) Cleanup(context.Context) error {
	*p.events = append(*p.events, p.id+":cleanup")
	return nil
}

type testView struct {
	events  *[]string
	results []Result
	mutate  bool
}

func (v *testView) Unavailable(string, string) {}
func (v *testView) Failure(Result)             {}
func (v *testView) DependencyNotice()          {}
func (v *testView) Interrupted()               {}
func (v *testView) Stage(string, string)       {}
func (v *testView) Plan(p Plan, _ bool) {
	*v.events = append(*v.events, p.Provider+":plan")
	if v.mutate && len(p.Candidates) > 0 {
		p.Candidates[0].ID = "render-mutation"
	}
}
func (v *testView) Confirm(context.Context) (bool, bool) {
	*v.events = append(*v.events, "confirm")
	return true, true
}
func (v *testView) Summary(r Result) { v.results = append(v.results, r) }

func TestCanonicalOrderAndSingleFinalPlan(t *testing.T) {
	var events []string
	providers := []Provider{}
	for _, id := range []string{"softwareupdate", "npm", "brew", "mas"} {
		providers = append(providers, &testProvider{id: id, events: &events, candidates: []Candidate{{ID: id + "-package"}}})
	}
	view := &testView{events: &events}
	if code := Run(context.Background(), Options{}, providers, view); code != 0 {
		t.Fatal(code)
	}
	want := []string{"brew:detect", "brew:discover", "mas:detect", "mas:discover", "npm:detect", "npm:discover", "softwareupdate:detect", "softwareupdate:discover", "brew:plan", "mas:plan", "npm:plan", "softwareupdate:plan", "confirm", "brew:apply", "mas:apply", "npm:apply", "softwareupdate:apply"}
	if !reflect.DeepEqual(events, want) {
		t.Fatal(events)
	}
	if providers[0].ID() != "softwareupdate" {
		t.Fatal("mutated caller registry")
	}
}
func TestDiscoveredPlanCannotBeBroadened(t *testing.T) {
	var events []string
	first := &testProvider{id: "brew", events: &events, candidates: []Candidate{{ID: "approved"}}}
	second := &testProvider{id: "mas", events: &events, candidates: []Candidate{{ID: "second"}}, discover: func() { first.candidates[0].ID = "late-change" }}
	view := &testView{events: &events, mutate: true}
	if code := Run(context.Background(), Options{Yes: true}, []Provider{first, second}, view); code != 0 {
		t.Fatal(code)
	}
	if first.applied[0].ID != "approved" || second.applied[0].ID != "second" {
		t.Fatal("approved plan changed", first.applied, second.applied)
	}
}
func TestInterruptionBeforeCleanupPreservesCompletedMutation(t *testing.T) {
	var events []string
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	p := &testProvider{id: "brew", events: &events, candidates: []Candidate{{ID: "approved"}}, apply: func() { cancel(process.SignalError{Signal: syscall.SIGTERM}) }}
	view := &testView{events: &events}
	if code := Run(ctx, Options{Yes: true, Clean: true}, []Provider{p}, view); code != 143 {
		t.Fatal(code)
	}
	result := view.results[0]
	if result.Status != "interrupted" || result.Stage != "cleanup" || !result.MutationAttempted || len(result.Outcome.Completed) != 1 {
		t.Fatal(result)
	}
	for _, event := range events {
		if event == "brew:cleanup" {
			t.Fatal("cleanup scheduled after signal")
		}
	}
}

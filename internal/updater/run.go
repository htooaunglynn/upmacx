package updater

import (
	"context"
	"fmt"
	"syscall"

	"github.com/htooaunglynn/upmacx/internal/process"
)

type Options struct{ Check, Yes, Clean, Explicit bool }

// View keeps presentation and terminal input outside orchestration/provider code.
type View interface {
	Unavailable(string, string)
	Plan(Plan, bool)
	Failure(Result)
	Confirm(context.Context) (bool, bool) // approved, terminal available
	DependencyNotice()
	Summary(Result)
	Interrupted()
	Stage(string, string)
}

func Run(ctx context.Context, o Options, providers []Provider, view View) int {
	providers = ordered(providers)
	results := make([]Result, len(providers))
	plans := make([]Plan, len(providers))
	work, failed := false, false
	fail := func(i int, stage string, err error) {
		results[i].Status = "failed"
		results[i].Stage = stage
		results[i].Err = err
		if ctx.Err() != nil {
			results[i].Status = "interrupted"
		}
		failed = true
		view.Failure(results[i])
	}
	for i, p := range providers {
		results[i].ID = p.ID()
		if ctx.Err() != nil {
			results[i].Status = "interrupted"
			continue
		}
		results[i].Stage = "detect"
		view.Stage(p.ID(), "detecting")
		avail, e := p.Detect(ctx)
		if e != nil {
			fail(i, "detect", e)
			continue
		}
		if !avail.Available {
			results[i].Status = "unavailable"
			view.Unavailable(p.ID(), avail.Reason)
			if o.Explicit {
				fail(i, "detect", fmt.Errorf("%s", avail.Reason))
			}
			continue
		}
		results[i].Stage = "discover"
		view.Stage(p.ID(), "refreshing metadata and checking updates")
		plan, e := p.Discover(ctx)
		if e != nil {
			fail(i, "discover", e)
			continue
		}
		plans[i] = clonePlan(plan)
		results[i].Status = "up-to-date"
		cleanup := o.Clean && p.ID() == "brew"
		if len(plan.Candidates) > 0 {
			if plan.ReportOnly {
				results[i].Status = "report-only"
			} else {
				results[i].Status = "checked"
				work = true
			}
		}
		if cleanup {
			work = true
		}
	}
	// Every provider finishes discovery before rendering the final consent plan.
	for i, p := range providers {
		if results[i].Status == "failed" || results[i].Status == "unavailable" || results[i].Status == "interrupted" {
			continue
		}
		view.Plan(clonePlan(plans[i]), o.Clean && p.ID() == "brew")
	}
	code := 0
	if !o.Check && work && ctx.Err() == nil {
		view.DependencyNotice()
		approved := o.Yes
		if !approved {
			var available bool
			approved, available = view.Confirm(ctx)
			if !available {
				code = 2
			}
		}
		for i, p := range providers {
			if results[i].Status == "failed" || results[i].Status == "unavailable" || plans[i].ReportOnly {
				continue
			}
			if ctx.Err() != nil {
				results[i].Status = "interrupted"
				continue
			}
			cleanup := o.Clean && p.ID() == "brew"
			if !approved {
				if len(plans[i].Candidates) > 0 || cleanup {
					results[i].Status = "declined"
				}
				continue
			}
			if len(plans[i].Candidates) > 0 {
				results[i].Stage = "apply"
				results[i].MutationAttempted = true
				view.Stage(p.ID(), "applying approved updates")
				out, e := p.Apply(ctx, clonePlan(plans[i]))
				results[i].Outcome = out
				if e != nil {
					fail(i, "apply (partial/unknown changes possible)", e)
					continue
				}
				results[i].Status = "succeeded"
			}
			if cleanup && ctx.Err() != nil {
				results[i].Status = "interrupted"
				results[i].Stage = "cleanup"
			}
			if cleanup && ctx.Err() == nil {
				results[i].Stage = "cleanup"
				results[i].MutationAttempted = true
				view.Stage(p.ID(), "running approved cleanup")
				if e := p.Cleanup(ctx); e != nil {
					fail(i, "cleanup", e)
				} else {
					results[i].Status = "succeeded"
				}
			}
		}
	}
	for i := range results {
		view.Summary(results[i])
	}
	if ctx.Err() != nil {
		view.Interrupted()
		if s, ok := context.Cause(ctx).(process.SignalError); ok && s.Signal == syscall.SIGTERM {
			return 143
		}
		return 130
	}
	if code != 0 {
		return code
	}
	if failed {
		return 1
	}
	return 0
}

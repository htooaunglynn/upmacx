package brew

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/htooaunglynn/upmacx/internal/process"
	"github.com/htooaunglynn/upmacx/internal/updater"
)

type Provider struct {
	runner process.Runner
	path   string
}

func New(r process.Runner) *Provider { return &Provider{runner: r} }
func (p *Provider) ID() string       { return "brew" }
func environment() map[string]string {
	return map[string]string{"LC_ALL": "C", "HOMEBREW_NO_INSTALL_CLEANUP": "1", "HOMEBREW_NO_AUTO_UPDATE": "1", "HOMEBREW_UPGRADE_GREEDY": "0", "HOMEBREW_UPGRADE_GREEDY_CASKS": "", "HOMEBREW_UPGRADE_GREEDY_LATEST": "0", "HOMEBREW_UPGRADE_GREEDY_AUTO_UPDATES": "0"}
}
func (p *Provider) run(ctx context.Context, mode string, args ...string) (process.CommandResult, error) {
	if p.path == "" {
		return process.CommandResult{}, errors.New("brew not detected")
	}
	stdin := "closed"
	if mode == "stream" {
		stdin = "inherit"
	}
	r, err := p.runner.Run(ctx, process.Command{Path: p.path, Args: args, Env: environment(), Mode: mode, Stdin: stdin})
	if err != nil && len(r.Stderr) > 0 {
		err = fmt.Errorf("%w: %s", err, strings.TrimSpace(string(r.Stderr)))
	}
	return r, err
}
func (p *Provider) Detect(ctx context.Context) (updater.Availability, error) {
	path, err := p.runner.LookPath("brew")
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) || errors.Is(err, os.ErrNotExist) {
			return updater.Availability{Reason: "brew is not installed; install Homebrew manually"}, nil
		}
		return updater.Availability{}, err
	}
	p.path = path
	r, err := p.run(ctx, "capture", "--version")
	if err != nil {
		return updater.Availability{}, err
	}
	version := strings.Split(strings.TrimSpace(string(r.Stdout)), "\n")[0]
	// This adapter's schema/environment semantics were reviewed against 7.0.10.
	if version != "Homebrew 7.0.10" {
		return updater.Availability{Version: version, Reason: "unverified Homebrew version (reviewed: 7.0.10)"}, nil
	}
	for _, probe := range []struct {
		command string
		tokens  []string
	}{{"outdated", []string{"--json", "v2", "--greedy"}}, {"list", []string{"--pinned", "--cask"}}, {"upgrade", []string{"--formula", "--cask"}}, {"cleanup", []string{"cleanup"}}} {
		r, e := p.run(ctx, "capture", "help", probe.command)
		if e != nil {
			// Unsupported help exits indicate a missing capability. Timeouts,
			// cancellation, output overflow and execution failures are runtime
			// failures; default selection must not silently skip them.
			if r.ExitCode <= 0 || errors.Is(e, context.Canceled) || errors.Is(e, context.DeadlineExceeded) || errors.Is(e, process.ErrOverflow) {
				return updater.Availability{}, fmt.Errorf("probe %s: %w", probe.command, e)
			}
			return updater.Availability{Version: version, Reason: "required help unavailable: " + probe.command}, nil
		}
		for _, s := range probe.tokens {
			if !strings.Contains(string(r.Stdout), s) {
				return updater.Availability{Version: version, Reason: "missing capability: " + s}, nil
			}
		}
	}
	return updater.Availability{Available: true, Version: version}, nil
}

var token = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9+_.@-]*(/[A-Za-z0-9][A-Za-z0-9+_.@-]*){0,2}$`)

type record struct {
	Name      string   `json:"name"`
	Installed []string `json:"installed_versions"`
	Target    string   `json:"current_version"`
	Pinned    *bool    `json:"pinned"`
}

func Parse(data []byte, pinned map[string]bool) (updater.Plan, error) {
	plan := updater.Plan{Provider: "brew"}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return plan, err
	}
	seen := map[string]bool{}
	for _, kind := range []string{"formula", "cask"} {
		key := "formulae"
		if kind == "cask" {
			key = "casks"
		}
		var records []record
		if len(raw[key]) == 0 || string(raw[key]) == "null" {
			return plan, fmt.Errorf("missing %s array", key)
		}
		if e := json.Unmarshal(raw[key], &records); e != nil {
			return plan, e
		}
		for _, r := range records {
			if !token.MatchString(r.Name) || len(r.Installed) == 0 || r.Target == "" || r.Pinned == nil {
				return updater.Plan{}, errors.New("incompatible or unsafe outdated record")
			}
			for _, v := range r.Installed {
				if v == "" {
					return updater.Plan{}, errors.New("missing installed version")
				}
			}
			id := kind + ":" + r.Name
			if seen[id] {
				return updater.Plan{}, errors.New("duplicate candidate")
			}
			seen[id] = true
			short := r.Name[strings.LastIndex(r.Name, "/")+1:]
			if *r.Pinned || pinned[r.Name] || pinned[short] {
				plan.Exclusions = append(plan.Exclusions, updater.Exclusion{Name: r.Name, Reason: "pinned"})
				continue
			}
			if r.Target == "latest" || r.Target == "HEAD" {
				plan.Exclusions = append(plan.Exclusions, updater.Exclusion{Name: r.Name, Reason: "unbounded target"})
				continue
			}
			plan.Candidates = append(plan.Candidates, updater.Candidate{ID: r.Name, Name: r.Name, Kind: kind, Installed: strings.Join(r.Installed, ", "), Target: r.Target})
		}
	}
	return plan, nil
}
func (p *Provider) Discover(ctx context.Context) (updater.Plan, error) {
	if _, e := p.run(ctx, "capture", "update"); e != nil {
		return updater.Plan{}, fmt.Errorf("refresh: %w", e)
	}
	r, e := p.run(ctx, "capture", "list", "--pinned")
	if e != nil {
		return updater.Plan{}, e
	}
	pinned := map[string]bool{}
	for _, line := range strings.Split(strings.TrimSpace(string(r.Stdout)), "\n") {
		if line == "" {
			continue
		}
		if !token.MatchString(line) {
			return updater.Plan{}, errors.New("malformed pinned list")
		}
		pinned[line] = true
	}
	r, e = p.run(ctx, "capture", "outdated", "--json=v2")
	if e != nil {
		return updater.Plan{}, e
	}
	return Parse(r.Stdout, pinned)
}
func (p *Provider) Apply(ctx context.Context, plan updater.Plan) (updater.Outcome, error) {
	out := updater.Outcome{Provider: "brew"}
	if plan.Provider != "brew" || plan.ReportOnly || len(plan.Candidates) == 0 {
		return out, errors.New("invalid or empty brew plan")
	}
	for _, c := range plan.Candidates {
		if !token.MatchString(c.ID) || (c.Kind != "formula" && c.Kind != "cask") {
			return out, errors.New("unsafe candidate")
		}
	}
	for _, c := range plan.Candidates {
		if ctx.Err() != nil {
			return out, ctx.Err()
		}
		if _, e := p.run(ctx, "stream", "upgrade", "--"+c.Kind, c.ID); e != nil {
			out.Status = "partial/unknown"
			return out, e
		}
		out.Completed = append(out.Completed, c.ID)
	}
	out.Status = "succeeded"
	return out, nil
}
func (p *Provider) Cleanup(ctx context.Context) error {
	_, e := p.run(ctx, "stream", "cleanup")
	return e
}

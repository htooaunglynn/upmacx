package cli

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/htooaunglynn/upmacx/internal/updater"
)

type App struct {
	Version, OS string
	UID         int
	In          io.Reader
	Out, Err    io.Writer
	Terminal    func() bool
	Providers   []updater.Provider
}

func (a App) Run(ctx context.Context, args []string) int {
	o, err := Parse(args)
	if err != nil {
		fmt.Fprintln(a.Err, err)
		return 2
	}
	if o.Help {
		fmt.Fprint(a.Out, Usage)
		return 0
	}
	if o.Version {
		fmt.Fprintln(a.Out, "upmacx "+a.Version)
		return 0
	}
	if a.OS != "darwin" || a.UID == 0 {
		fmt.Fprintln(a.Err, "execution requires macOS and a non-root user")
		return 1
	}
	return updater.Run(ctx, updater.Options{Check: o.Check, Yes: o.Yes, Clean: o.Clean, Explicit: o.Explicit}, a.Providers, a)
}
func (a App) Unavailable(id, reason string) {
	fmt.Fprintf(a.Out, "[%s] unavailable: %s\n", id, Safe(reason))
}
func (a App) Plan(p updater.Plan, cleanup bool) {
	for _, c := range p.Candidates {
		fmt.Fprintf(a.Out, "[%s] %s %s: %s -> %s %s\n", p.Provider, Safe(c.Kind), Safe(c.Name), Safe(c.Installed), Safe(c.Target), Safe(c.Note))
	}
	for _, x := range p.Exclusions {
		fmt.Fprintf(a.Out, "[%s] excluded %s: %s\n", p.Provider, Safe(x.Name), Safe(x.Reason))
	}
	if cleanup {
		fmt.Fprintln(a.Out, "[brew] planned cleanup")
	}
}
func (a App) Failure(r updater.Result) {
	fmt.Fprintf(a.Err, "[%s] %s: %s\n", r.ID, r.Stage, Safe(r.Err.Error()))
}
func (a App) DependencyNotice() {
	fmt.Fprintln(a.Out, "Package managers may resolve dependency changes beyond the listed candidates; target versions are previews.")
}
func (a App) Confirm(ctx context.Context) (bool, bool) {
	if a.Terminal == nil || !a.Terminal() {
		fmt.Fprintln(a.Err, "confirmation required: use a terminal or --yes")
		return false, false
	}
	fmt.Fprint(a.Out, "Continue with planned changes? [y/N]: ")
	line, e := confirmationLine(ctx, a.In)
	line = strings.ToLower(strings.TrimSpace(line))
	return e == nil && (line == "y" || line == "yes"), true
}
func (a App) Summary(r updater.Result) {
	fmt.Fprintf(a.Out, "[%s] %s", r.ID, r.Status)
	if len(r.Outcome.Completed) > 0 {
		fmt.Fprintf(a.Out, " (completed: %s)", Safe(strings.Join(r.Outcome.Completed, ", ")))
	}
	fmt.Fprintln(a.Out)
}
func (a App) Interrupted() {
	fmt.Fprintln(a.Err, "Interrupted installations may require package-manager repair; no repair was attempted.")
}

// A one-byte reader avoids consuming input intended for a child installer.
func confirmationLine(ctx context.Context, in io.Reader) (string, error) {
	type answer struct {
		line string
		err  error
	}
	ch := make(chan answer, 1)
	go func() {
		var b strings.Builder
		var one [1]byte
		for {
			n, e := in.Read(one[:])
			if n > 0 {
				if one[0] == '\n' {
					ch <- answer{line: b.String()}
					return
				}
				if b.Len() < 4096 {
					b.WriteByte(one[0])
				} else {
					ch <- answer{err: fmt.Errorf("confirmation too long")}
					return
				}
			}
			if e != nil {
				ch <- answer{err: e}
				return
			}
		}
	}()
	select {
	case a := <-ch:
		return a.line, a.err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

package cli

import (
	"errors"
	"flag"
	"io"
)

type Options struct{ Check, Yes, Clean, System, Help, Version, Explicit bool }
type selection struct {
	set   bool
	value string
}

func (s *selection) String() string { return s.value }
func (s *selection) Set(v string) error {
	if s.set {
		return errors.New("repeated --providers")
	}
	s.set = true
	s.value = v
	return nil
}
func Parse(args []string) (Options, error) {
	var o Options
	var p selection
	f := flag.NewFlagSet("upmacx", flag.ContinueOnError)
	f.SetOutput(io.Discard)
	f.BoolVar(&o.Check, "check", false, "")
	f.BoolVar(&o.Yes, "yes", false, "")
	f.BoolVar(&o.Clean, "clean", false, "")
	f.BoolVar(&o.System, "system", false, "")
	f.BoolVar(&o.Help, "help", false, "")
	f.BoolVar(&o.Help, "h", false, "")
	f.BoolVar(&o.Version, "version", false, "")
	f.Var(&p, "providers", "")
	if err := f.Parse(args); err != nil {
		return o, err
	}
	if f.NArg() != 0 {
		return o, errors.New("positional arguments are unsupported")
	}
	o.Explicit = p.set
	if p.set {
		for _, id := range split(p.value) {
			if id != "brew" {
				return o, errors.New("unavailable provider ID: " + id + " (this build supports brew)")
			}
		}
	}
	if o.Help || o.Version {
		return o, nil
	}
	if o.Check && (o.Clean || o.System) {
		return o, errors.New("--check cannot be combined with --clean or --system")
	}
	if o.System {
		return o, errors.New("--system is unavailable until the softwareupdate provider ships")
	}
	return o, nil
}

const Usage = `upmacx — Your macOS update assistant
Usage: upmacx [--check] [--yes] [--clean] [--providers brew]
  --check       Refresh metadata and inspect; never install or clean
  --yes         Skip upmacx confirmation only
  --clean       Request Homebrew cleanup after a successful stage
  --providers   Comma-separated provider IDs (shipped: brew)
  --help, -h    Show help without running providers
  --version     Show version without running providers
Only Homebrew formulae and eligible casks are covered. System updates and
other package managers are unavailable. Cleanup is opt-in. No automatic
privilege escalation, dependency installation, or reboot.
`

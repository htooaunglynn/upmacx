package updater

import "context"

type Availability struct {
	Available       bool
	Version, Reason string
}
type Candidate struct{ ID, Name, Installed, Target, Kind, Note string }
type Exclusion struct{ Name, Reason string }
type Plan struct {
	Provider   string
	Candidates []Candidate
	Exclusions []Exclusion
	ReportOnly bool
}
type Outcome struct {
	Provider, Status string
	Completed        []string
	Notes            []string
}
type Provider interface {
	ID() string
	Detect(context.Context) (Availability, error)
	Discover(context.Context) (Plan, error)
	Apply(context.Context, Plan) (Outcome, error)
	Cleanup(context.Context) error
}
type Result struct {
	ID, Status, Stage string
	Err               error
	Outcome           Outcome
	MutationAttempted bool
}

package updater

import (
	"slices"
	"sort"
)

// ordered copies the injected registry and enforces the execution contract.
// Unknown test-only providers retain registration order after known providers.
func ordered(providers []Provider) []Provider {
	copy := slices.Clone(providers)
	rank := func(id string) int {
		for i, known := range []string{"brew", "mas", "npm", "pnpm", "pipx", "cargo", "softwareupdate"} {
			if id == known {
				return i
			}
		}
		return 7
	}
	sort.SliceStable(copy, func(i, j int) bool { return rank(copy[i].ID()) < rank(copy[j].ID()) })
	return copy
}
func clonePlan(p Plan) Plan {
	p.Candidates = slices.Clone(p.Candidates)
	p.Exclusions = slices.Clone(p.Exclusions)
	return p
}

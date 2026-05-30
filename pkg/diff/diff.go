package diff

import (
	"fmt"
	"sort"
)

// DiffMaps returns human-friendly diffs comparing a and b.
func Maps(a, b map[string][]byte) []string {
	keys := map[string]struct{}{}
	for k := range a {
		keys[k] = struct{}{}
	}
	for k := range b {
		keys[k] = struct{}{}
	}
	list := make([]string, 0, len(keys))
	for k := range keys {
		list = append(list, k)
	}
	sort.Strings(list)
	out := []string{}
	for _, k := range list {
		av, aok := a[k]
		bv, bok := b[k]
		switch {
		case !aok && bok:
			out = append(out, fmt.Sprintf("+ %s", k))
		case aok && !bok:
			out = append(out, fmt.Sprintf("- %s", k))
		default:
			if string(av) != string(bv) {
				out = append(out, fmt.Sprintf("~ %s", k))
			} else {
				out = append(out, fmt.Sprintf("  %s", k))
			}
		}
	}
	return out
}

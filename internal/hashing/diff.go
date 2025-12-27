package hashing

import (
	"sort"
)

// Diff compares two states and returns the keys of sections that have changed.
// A section is considered changed if its value in newState is different from oldState,
// or if it exists in newState but not in oldState.
func Diff(oldState, newState map[string]interface{}) ([]string, error) {
	changed := []string{}

	// Check for changed or new sections in newState
	for k, vNew := range newState {
		vOld, exists := oldState[k]
		if !exists {
			changed = append(changed, k)
			continue
		}

		// Use Hash to compare values accurately
		hOld, err := Hash(vOld)
		if err != nil {
			return nil, err
		}
		hNew, err := Hash(vNew)
		if err != nil {
			return nil, err
		}

		if hOld != hNew {
			changed = append(changed, k)
		}
	}

	// Sort results for determinism
	sort.Strings(changed)

	return changed, nil
}

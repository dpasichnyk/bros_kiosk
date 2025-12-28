package hashing

import (
	"sort"
)

func Diff(oldState, newState map[string]interface{}) ([]string, error) {
	changed := []string{}

	for k, vNew := range newState {
		vOld, exists := oldState[k]
		if !exists {
			changed = append(changed, k)
			continue
		}

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

	sort.Strings(changed)

	return changed, nil
}

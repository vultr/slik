package str

import "strings"

// SplitNoEmpty splits string based on separator and removes empties
func SplitNoEmpty(s, sep string) []string {
	e := strings.Split(s, sep)

	var r []string
	for i := range e {
		if strings.TrimSpace(e[i]) != "" {
			r = append(r, e[i])
		}
	}

	return r
}

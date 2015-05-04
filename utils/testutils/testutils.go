package testutils

import "unicode"

func SpacesRemoved(s string) string {
	ret := ""
	for _, r := range []rune(s) {
		if !unicode.IsSpace(r) {
			ret += string(r)
		}
	}

	return ret
}

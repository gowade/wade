package wade

import (
	"strings"
)

func defaultHelpers() map[string]interface{} {
	return map[string]interface{}{
		"toUpper": strings.ToUpper,
		"toLower": strings.ToLower,
		"concat": func(s1, s2 string) string {
			return s1 + s2
		},
	}
}

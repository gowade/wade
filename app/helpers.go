package app

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/phaikawl/wade/utils"
)

var defaultHelpers = map[string]interface{}{
	"toUpper": strings.ToUpper,

	"toLower": strings.ToLower,

	"join": func(s1 string, strings ...string) string {
		for _, s := range strings {
			s1 += s
		}
		return s1
	},

	"format": func(format string, values ...interface{}) string {
		return fmt.Sprintf(format, values...)
	},

	"eq": func(a, b interface{}) bool {
		return reflect.DeepEqual(a, b)
	},

	"not": func(a bool) bool {
		return !a
	},

	"empty": func(collection interface{}) bool {
		return reflect.ValueOf(collection).Len() == 0
	},

	"len": func(collection interface{}) int {
		return reflect.ValueOf(collection).Len()
	},

	"emptyStr": func(str string) bool {
		return str == ""
	},

	"str": utils.ToString,
}

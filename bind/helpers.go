package bind

import (
	"fmt"
	"reflect"
	"strings"
)

type PageManager interface {
	CurrentPageId() string
	PageUrl(string, ...interface{}) (string, error)
	Fullpath(string) string
}

type UrlInfo struct {
	path    string
	fullUrl string
}

func RegisterInternalHelpers(pm PageManager, b *Binding) {
	b.RegisterHelper("url", func(pageid string, params ...interface{}) string {
		url, err := pm.PageUrl(pageid, params...)
		if err != nil {
			panic(fmt.Errorf(`url helper error: "%v", when getting url for page "%v"`, err.Error(), pageid))
		}
		return url
	})
}

func defaultHelpers() map[string]interface{} {
	return map[string]interface{}{
		"toUpper": strings.ToUpper,

		"toLower": strings.ToLower,

		"concat": func(s1, s2 string) string {
			return s1 + s2
		},

		"isEqual": func(a, b interface{}) bool {
			return reflect.DeepEqual(a, b)
		},

		"not": func(a bool) bool {
			return !a
		},

		"isEmpty": func(collection interface{}) bool {
			return reflect.ValueOf(collection).Len() == 0
		},

		"len": func(collection interface{}) int {
			return reflect.ValueOf(collection).Len()
		},

		"isEmptyStr": func(str string) bool {
			return str == ""
		},

		"toStr": toString,
	}
}

package bind

import (
	"fmt"
	"reflect"
	"strings"
)

type PageManager interface {
	PageUrl(string, []interface{}) (string, error)
	Url(string) string
}

type UrlInfo struct {
	path    string
	fullUrl string
}

func RegisterUrlHelper(pm PageManager, b *Binding) {
	b.RegisterHelper("url", func(pageid string, params ...interface{}) UrlInfo {
		url, err := pm.PageUrl(pageid, params)
		if err != nil {
			panic(fmt.Errorf(`url helper error: "%v", when getting url for page "%v"`, err.Error(), pageid))
		}
		return UrlInfo{url, pm.Url(url)}
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
			return reflect.DeepEqual(reflect.TypeOf(a), reflect.TypeOf(b))
		},
		"isEmpty": func(collection interface{}) bool {
			return reflect.ValueOf(collection).Len() == 0
		},
		"len": func(collection interface{}) int {
			return reflect.ValueOf(collection).Len()
		},
	}
}

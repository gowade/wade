package utils

import (
	"fmt"
	neturl "net/url"
	"reflect"
	"unicode"
)

// NoSp returns returns the string without space characters (space, newline, tabs...)
func NoSp(src string) string {
	r := ""
	for _, c := range src {
		if !unicode.IsSpace(c) {
			r += string(c)
		}
	}

	return r
}

func ToString(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%v", value)
}

type M map[string]interface{}

// UrlQuery adds query arguments (?arg1=value1&arg2=value2...)
// specified in the given name-value map to a given url and returns the new result
func UrlQuery(url string, args M) string {
	m := make(map[string][]string)
	for k, v := range args {
		m[k] = []string{ToString(v)}
	}

	qs := neturl.Values(m).Encode()
	if qs == "" {
		return url
	}

	return url + "?" + qs
}

type MapItem struct {
	Key string
	Val interface{}
}

type byStringKey []MapItem

func (a byStringKey) Len() int           { return len(a) }
func (a byStringKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byStringKey) Less(i, j int) bool { return a[i].Key > a[j].Key }

// ListFromMap returns a list of MapItem sorted in alphabetical order from a map[string]T.
// Insteaded for use when creating a scope map.
func ListFromMap(m interface{}) []MapItem {
	mv := reflect.ValueOf(m)
	list := make([]MapItem, mv.Len())
	for i, item := range mv.MapKeys() {
		list[i] = MapItem{
			Key: item.String(),
			Val: mv.MapIndex(item).Interface(),
		}
	}

	return list
}

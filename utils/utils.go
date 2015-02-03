package utils

import (
	"fmt"
	neturl "net/url"
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

	return fmt.Sprint(value)
}

type Map map[string]interface{}

// UrlQuery adds query arguments (?arg1=value1&arg2=value2...)
// specified in the given name-value map to a given url and returns the new one
func UrlQuery(url string, args Map) string {
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

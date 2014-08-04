package wade

import (
	"unicode"

	jq "github.com/gopherjs/jquery"
)

func ToElemSlice(elems jq.JQuery) []jq.JQuery {
	list := make([]jq.JQuery, elems.Length)
	elems.Each(func(i int, elem jq.JQuery) {
		list[i] = elem
	})

	return list
}

func camelize(src string) string {
	res := []rune{}
	startW := true
	for _, c := range src {
		if c == '-' {
			startW = true
			continue
		}
		ch := c
		if startW {
			ch = unicode.ToUpper(c)
			startW = false
		}
		res = append(res, ch)
	}
	return string(res)
}

type UrlInfo struct {
	path    string
	fullUrl string
}

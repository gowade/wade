package wade

import (
	"unicode"

	"github.com/phaikawl/wade/icommon"
)

var (
	IsWrapperElem = icommon.IsWrapperElem
)

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

type GetSetable interface {
	Get(key string) (v interface{}, ok bool)
	Set(key string, v interface{})
}

type Storage struct {
	GetSetable
}

func (s *Storage) GetBool(key string) (v bool, ok bool) {
	var ov interface{}
	ov, ok = s.Get(key)
	if ok {
		v = ov.(bool)
	}

	return
}

func (s *Storage) GetInt(key string) (v int, ok bool) {
	var ov interface{}
	ov, ok = s.Get(key)
	if ok {
		v = ov.(int)
	}

	return
}

func (s *Storage) GetStr(key string) (v string, ok bool) {
	var ov interface{}
	ov, ok = s.Get(key)
	if ok {
		v = ov.(string)
	}

	return
}

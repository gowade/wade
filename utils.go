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
	Get(key string, v interface{}) (ok bool)
	Set(key string, v interface{})
	Delete(key string)
}

type Storage struct {
	GetSetable
}

func (stg Storage) GetBool(key string) (v bool, ok bool) {
	ok = stg.Get(key, &v)
	return
}

func (stg Storage) GetStr(key string) (v string, ok bool) {
	ok = stg.Get(key, &v)
	return
}

func (stg Storage) GetInt(key string) (v int, ok bool) {
	ok = stg.Get(key, &v)
	return
}

//Get the stored value with key key and store it in v.
//Typically used for struct values.
func (stg Storage) GetTo(key string, v interface{}) bool {
	return stg.Get(key, v)
}

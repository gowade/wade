package wade

import (
	neturl "net/url"

	"github.com/phaikawl/wade/icommon"
	"github.com/phaikawl/wade/libs/http"
)

var (
	IsWrapperElem = icommon.IsWrapperElem
)

func UrlQuery(url string, args map[string][]string) string {
	qs := neturl.Values(args).Encode()
	if qs == "" {
		return url
	}

	return url + "?" + qs
}

func Http() *http.Client {
	return http.DefaultClient()
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

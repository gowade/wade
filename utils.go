package wade

import (
	"reflect"
	"unicode"

	"github.com/phaikawl/wade/services/http"
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

// SendFormTo sends a "form" with the specified data to a specified url and decode
// the validation errors to valdErrs, valdErrs must be a pointer
func SendFormTo(url string, data interface{}, valdErrs interface{}) *http.Response {
	if reflect.TypeOf(valdErrs).Kind() != reflect.Ptr {
		panic("valErrs target argument must be a pointer")
	}
	req := http.Service().NewRequest(http.MethodPost, url)
	req.SetData(data)
	r := req.DoSync()
	err := r.DecodeDataTo(valdErrs)
	if err != nil {
		panic(err.Error())
	}
	return r
}

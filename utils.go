package wade

import (
	"unicode"

	"github.com/phaikawl/wade/lib"
	"github.com/phaikawl/wade/services/http"
)

type FormResp lib.FormResp

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

func SendFormTo(url string, data interface{}, valdErrs *Validated) *Promise {
	req := http.Service().NewRequest(http.MethodPost, url)
	req.SetData(data)
	promise := NewPromise(valdErrs, req.DoAsync())
	promise.OnSuccess(func(r *http.Response) ModelUpdater {
		err := r.DecodeDataTo(&valdErrs.Errors)
		if err != nil {
			panic(err.Error())
		}
		return nil
	})

	return promise
}

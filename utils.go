package wade

import (
	"encoding/json"

	"github.com/gopherjs/gopherjs/js"
	"github.com/phaikawl/wade/lib"
	"github.com/phaikawl/wade/services/http"
)

type FormResp lib.FormResp

func SendFormTo(url string, data interface{}, valdErrs js.Object) *Promise {
	req := http.Service().NewRequest(http.MethodPost, url)
	req.SetData(data)
	promise := NewPromise(valdErrs, req.DoAsync())
	promise.OnSuccess(func(r *http.Response) ModelUpdater {
		ve := new(map[string]map[string]interface{})
		err := json.Unmarshal([]byte(`{"Password":{"minChar":"too short, minimum 6 characters"}}`), &ve)
		valdErrs = js.Global.Call("createObj")
		//println(valdErrs)
		if err != nil {
			panic(err.Error())
		}
		return nil
	})

	return promise
}

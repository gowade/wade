package clientside

import (
	"encoding/json"
	"reflect"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"github.com/phaikawl/wade"
	"github.com/phaikawl/wade/dom"
	jqdom "github.com/phaikawl/wade/dom/jquery"
	"github.com/phaikawl/wade/libs/http"
	xhr "github.com/phaikawl/wade/libs/http/clientside"
)

var (
	gJQ               = jquery.NewJQuery
	gGlobal js.Object = js.Global
)

func RenderBackend() wade.RenderBackend {
	doc := jqdom.Document()
	return wade.RenderBackend{
		JsBackend: &JsBackend{
			history: History{js.Global.Get("history")},
		},
		Document:    doc,
		HttpBackend: newCachedHttpBackend(xhr.XhrBackend{}, doc),
	}
}

type (
	JsBackend struct {
		history History
	}

	storage struct {
		js.Object
	}

	cachedHttpBackend struct {
		http.Backend
		cache map[string]concreteRecord
	}

	headers struct {
		Header http.HttpHeader
	}

	concreteResponse struct {
		http.Response
		Headers headers
	}

	concreteRecord struct {
		Response *concreteResponse
		http.HttpRecord
	}
)

func newCachedHttpBackend(backend http.Backend, doc dom.Selection) *cachedHttpBackend {
	b := &cachedHttpBackend{backend, make(map[string]concreteRecord)}
	sn := doc.Find("script[type='text/wadehttp']")
	if sn.Length() > 0 {
		cc := sn.Text()
		if cc != "" {
			err := json.Unmarshal([]byte(cc), &b.cache)
			if err != nil {
				panic(err.Error())
			}
		}
	}

	return b
}

func (c *cachedHttpBackend) Do(r *http.Request) (err error) {
	if record, ok := c.cache[http.RequestIdent(r)]; ok {
		err = record.Error
		r.Response = &record.Response.Response
	} else {
		//gopherjs:blocking
		err = c.Backend.Do(r)
	}

	return
}

func (stg storage) Get(key string, outVal interface{}) (ok bool) {
	jsv := stg.Object.Call("getItem", key)
	ok = !jsv.IsNull() && !jsv.IsUndefined()
	if ok {
		gv := jsv.Str()
		err := json.Unmarshal([]byte(gv), &outVal)
		if err != nil {
			panic(err.Error())
		}
	}
	return
}

func (stg storage) Set(key string, v interface{}) {
	s, err := json.Marshal(v)
	if err != nil {
		panic(err.Error())
	}
	stg.Object.Set(key, string(s))
}

func (stg storage) Delete(key string) {
	stg.Object.Delete(key)
}

// CheckJsDep checks if given js name exists
func (b *JsBackend) CheckJsDep(symbol string) bool {
	if gGlobal.Get(symbol).IsUndefined() {
		return false
	}

	return true
}

func (b *JsBackend) History() wade.History {
	return b.history
}

func (b *JsBackend) Watch(modelRefl reflect.Value, field string, callback func()) {
	obj := js.InternalObject(modelRefl.Interface()).Get("$val")
	rcb := func(prop string, action string,
		_ js.Object,
		_2 js.Object) {
		callback()
	}

	js.Global.Call("watch",
		obj,
		field,
		rcb)
}

func (b *JsBackend) WebStorages() (wade.Storage, wade.Storage) {
	return wade.Storage{storage{js.Global.Get("localStorage")}},
		wade.Storage{storage{js.Global.Get("sessionStorage")}}
}

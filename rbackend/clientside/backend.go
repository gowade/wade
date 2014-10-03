package clientside

import (
	"encoding/json"
	"reflect"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"github.com/phaikawl/wade"
	"github.com/phaikawl/wade/bind"
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

	observeCloser struct {
		osvs []js.Object
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

func (c observeCloser) Close() {
	for _, osv := range c.osvs {
		osv.Call("close")
	}
}

func jso(object reflect.Value) js.Object {
	return js.InternalObject(object.Interface()).Get("$val")
}

func (b *JsBackend) Watch(ctl bind.WatchCtl, callback bind.WatchCallback) bind.WatchCloser {
	cbWrap := func() {
		go callback(0, nil)
	}

	var osvs []js.Object
	switch ctl.FieldRefl.Kind() {
	case reflect.Slice:
		fn := func(fieldRefl reflect.Value) js.Object {
			o1 := js.Global.Get("ArrayObserver").New(jso(fieldRefl).Get("$array"))
			o1.Call("open", cbWrap)

			return o1
		}

		o2 := js.Global.Get("PathObserver").New(jso(ctl.ModelRefl), ctl.Field)
		fn2 := func() {
			rf := ctl.NewFieldRefl()
			ctl.WatchAdd(rf, observeCloser{[]js.Object{fn(rf)}}, callback)

			go callback(ctl.FieldRefl.UnsafeAddr(), rf.Interface())
		}

		o2.Call("open", fn2)

		osvs = []js.Object{fn(ctl.FieldRefl), o2}

	case reflect.Map:
		osv := js.Global.Get("ObjectObserver").New(jso(ctl.FieldRefl))
		osv.Call("open", cbWrap)
		osvs = []js.Object{osv}

	default:
		osv := js.Global.Get("PathObserver").New(jso(ctl.ModelRefl), ctl.Field)
		osv.Call("open", cbWrap)
		osvs = []js.Object{osv}
	}

	return observeCloser{osvs}
}

func (b *JsBackend) DigestAll() {
	js.Global.Get("Platform").Call("performMicrotaskCheckpoint")
}

func (b *JsBackend) WebStorages() (wade.Storage, wade.Storage) {
	return wade.Storage{storage{js.Global.Get("localStorage")}},
		wade.Storage{storage{js.Global.Get("sessionStorage")}}
}

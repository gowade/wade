package clientside

import (
	"reflect"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"github.com/phaikawl/wade"
	jqdom "github.com/phaikawl/wade/dom/jquery"
	xhr "github.com/phaikawl/wade/libs/http/clientside"
)

var (
	gJQ               = jquery.NewJQuery
	gGlobal js.Object = js.Global
)

func RenderBackend() wade.RenderBackend {
	return wade.RenderBackend{
		JsBackend: &JsBackend{
			history: History{js.Global.Get("history")},
		},
		Document:    jqdom.Document(),
		HttpBackend: xhr.XhrBackend{},
	}
}

type (
	JsBackend struct {
		history History
	}

	storage struct {
		js.Object
	}
)

func (s storage) Get(key string) (v interface{}, ok bool) {
	val := s.Object.Get(key)
	if !val.IsUndefined() {
		v = val.Interface()
		ok = true
		return
	}

	return
}

// CheckJsDep checks if given js name exists
func (b *JsBackend) CheckJsDep(symbol string) bool {
	if gGlobal.Get(symbol).IsUndefined() {
		return false
	}

	return true
}

// Watch calls Watch.js to watch the object's changes
func (b *JsBackend) Watch(modelRefl reflect.Value, field string, callback func()) {
	obj := js.InternalObject(modelRefl.Interface()).Get("$val")
	js.Global.Call("watch",
		obj,
		field,
		func(prop string, action string,
			_ js.Object,
			_2 js.Object) {
			callback()
		})
}

func (b *JsBackend) History() wade.History {
	return b.history
}

func (b *JsBackend) WebStorages() (wade.Storage, wade.Storage) {
	return wade.Storage{storage{js.Global.Get("localStorage")}},
		wade.Storage{storage{js.Global.Get("sessionStorage")}}
}

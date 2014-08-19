package jsbackend

import (
	"reflect"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
)

var (
	gJQ                = jquery.NewJQuery
	gGlobal  js.Object = js.Global
	gBackend *BackendImp
)

func Get() *BackendImp {
	if gBackend == nil {
		gBackend = &BackendImp{
			History: History{js.Global.Get("history")},
		}
	}
	return gBackend
}

type BackendImp struct {
	History History
}

// CheckJsDep checks if given js name exists
func (b *BackendImp) CheckJsDep(symbol string) bool {
	if gGlobal.Get(symbol).IsUndefined() {
		return false
	}

	return true
}

// Watch calls Watch.js to watch the object's changes
func (b *BackendImp) Watch(modelRefl reflect.Value, field string, callback func()) {
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

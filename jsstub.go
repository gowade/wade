package wade

import (
	"reflect"

	"github.com/gopherjs/gopherjs/js"
)

type (
	JsStub struct {
		Defined bool
		Value   interface{}
		Map     map[string]*JsStub
	}

	noopHistory struct {
		path  string
		title string
	}
)

func NewStubJsValue(value interface{}) *JsStub {
	return &JsStub{
		Value:   value,
		Defined: true,
		Map:     make(map[string]*JsStub),
	}
}

func (j *JsStub) Get(name string) js.Object {
	c, ok := j.Map[name]
	if !ok {
		return &JsStub{
			Defined: false,
		}
	}

	return c
}

func (j *JsStub) Set(name string, value interface{}) {
	j.Map[name] = NewStubJsValue(value)
}

func (j *JsStub) Delete(name string) {
	delete(j.Map, name)
}

func (j *JsStub) Length() int {
	if !j.Defined {
		return 0
	}
	return reflect.ValueOf(j.Value).Len()
}

func (j *JsStub) Index(i int) js.Object {
	if !j.Defined {
		return &JsStub{Defined: false}
	}
	v := reflect.ValueOf(j.Value).Index(i)
	return &JsStub{
		Value:   v,
		Defined: v.IsValid(),
		Map:     make(map[string]*JsStub),
	}
}

func (j *JsStub) SetIndex(i int, value interface{}) {
	if !j.Defined {
		return
	}
	reflect.ValueOf(j.Value).Index(i).Set(reflect.ValueOf(value))
}

func (j *JsStub) Call(name string, args ...interface{}) js.Object {
	if !j.Defined {
		return &JsStub{Defined: false}
	}

	params := make([]reflect.Value, len(args))
	for i, _ := range args {
		params[i] = reflect.ValueOf(args[i])
	}

	return NewStubJsValue(reflect.ValueOf(j.Map[name].Value).Call(params))
}

func (j *JsStub) Invoke(args ...interface{}) js.Object {
	if !j.Defined {
		return &JsStub{Defined: false}
	}
	params := make([]reflect.Value, len(args))
	for i, _ := range args {
		params[i] = reflect.ValueOf(args[i])
	}

	return NewStubJsValue(reflect.ValueOf(j.Value).Call(params))
}

func (j *JsStub) New(args ...interface{}) js.Object {
	if !j.Defined {
		return &JsStub{Defined: false}
	}

	v := NewStubJsValue(j.Value)
	v.Invoke(args...)
	return v
}

func (j *JsStub) Bool() bool {
	if !j.Defined {
		return false
	}
	return j.Value.(bool)
}

func (j *JsStub) Str() string {
	if !j.Defined {
		return ""
	}
	return j.Value.(string)
}

func (j *JsStub) Int() int {
	if !j.Defined {
		return 0
	}
	return j.Value.(int)
}

func (j *JsStub) Int64() int64 {
	return j.Value.(int64)
}

func (j *JsStub) Uint64() uint64 {
	if !j.Defined {
		return 0
	}
	return j.Value.(uint64)
}

func (j *JsStub) Float() float64 {
	if !j.Defined {
		return 0
	}
	return j.Value.(float64)
}

func (j *JsStub) Interface() interface{} {
	if !j.Defined {
		return nil
	}
	return j.Value
}

func (j *JsStub) Unsafe() uintptr {
	return reflect.ValueOf(j.Value).UnsafeAddr()
}

func (j *JsStub) IsUndefined() bool {
	return !j.Defined
}

func (j *JsStub) IsNull() bool {
	return j.Value == nil
}

func NewNoopHistory() *noopHistory {
	return &noopHistory{path: "/"}
}

func (h *noopHistory) ReplaceState(title string, path string) {
	h.path = path
	h.title = title
}

func (h *noopHistory) PushState(title string, path string) {
	h.path = path
	h.title = title
}

func (h *noopHistory) OnPopState(fn func()) {}

func (h *noopHistory) CurrentPath() string {
	return h.path
}

func (h *noopHistory) RedirectTo(url string) {
}

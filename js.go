package wade

import (
	"reflect"

	"github.com/gopherjs/gopherjs/js"
)

type (
	jsStub struct {
		Defined bool
		Value   interface{}
		Map     map[string]*jsStub
	}

	noopHistory struct {
		path  string
		title string
	}
)

func newStubJsValue(value interface{}) *jsStub {
	return &jsStub{
		Value:   value,
		Defined: true,
		Map:     make(map[string]*jsStub),
	}
}

func (j *jsStub) Get(name string) js.Object {
	c, ok := j.Map[name]
	if !ok {
		return &jsStub{
			Defined: false,
		}
	}

	return c
}

func (j *jsStub) Set(name string, value interface{}) {
	j.Map[name] = newStubJsValue(value)
}

func (j *jsStub) Delete(name string) {
	delete(j.Map, name)
}

func (j *jsStub) Length() int {
	if !j.Defined {
		return 0
	}
	return reflect.ValueOf(j.Value).Len()
}

func (j *jsStub) Index(i int) js.Object {
	if !j.Defined {
		return &jsStub{Defined: false}
	}
	v := reflect.ValueOf(j.Value).Index(i)
	return &jsStub{
		Value:   v,
		Defined: v.IsValid(),
		Map:     make(map[string]*jsStub),
	}
}

func (j *jsStub) SetIndex(i int, value interface{}) {
	if !j.Defined {
		return
	}
	reflect.ValueOf(j.Value).Index(i).Set(reflect.ValueOf(value))
}

func (j *jsStub) Call(name string, args ...interface{}) js.Object {
	if !j.Defined {
		return &jsStub{Defined: false}
	}

	params := make([]reflect.Value, len(args))
	for i, _ := range args {
		params[i] = reflect.ValueOf(args[i])
	}

	return newStubJsValue(reflect.ValueOf(j.Map[name].Value).Call(params))
}

func (j *jsStub) Invoke(args ...interface{}) js.Object {
	if !j.Defined {
		return &jsStub{Defined: false}
	}
	params := make([]reflect.Value, len(args))
	for i, _ := range args {
		params[i] = reflect.ValueOf(args[i])
	}

	return newStubJsValue(reflect.ValueOf(j.Value).Call(params))
}

func (j *jsStub) New(args ...interface{}) js.Object {
	if !j.Defined {
		return &jsStub{Defined: false}
	}

	v := newStubJsValue(j.Value)
	v.Invoke(args...)
	return v
}

func (j *jsStub) Bool() bool {
	if !j.Defined {
		return false
	}
	return j.Value.(bool)
}

func (j *jsStub) Str() string {
	if !j.Defined {
		return ""
	}
	return j.Value.(string)
}

func (j *jsStub) Int() int {
	if !j.Defined {
		return 0
	}
	return j.Value.(int)
}

func (j *jsStub) Int64() int64 {
	return j.Value.(int64)
}

func (j *jsStub) Uint64() uint64 {
	if !j.Defined {
		return 0
	}
	return j.Value.(uint64)
}

func (j *jsStub) Float() float64 {
	if !j.Defined {
		return 0
	}
	return j.Value.(float64)
}

func (j *jsStub) Interface() interface{} {
	if !j.Defined {
		return nil
	}
	return j.Value
}

func (j *jsStub) Unsafe() uintptr {
	return reflect.ValueOf(j.Value).UnsafeAddr()
}

func (j *jsStub) IsUndefined() bool {
	return !j.Defined
}

func (j *jsStub) IsNull() bool {
	return j.Value == nil
}

func NewNoopHistory(path string) *noopHistory {
	return &noopHistory{path: path}
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

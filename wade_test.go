package wade

import (
	"testing"

	"github.com/gopherjs/gopherjs/js"
	"github.com/phaikawl/wade/services/http"
)

type JsStub struct{}

func (j *JsStub) Get(name string) js.Object {
	return j
}
func (j *JsStub) Set(name string, value interface{}) {
}
func (j *JsStub) Delete(name string) {
}
func (j *JsStub) Length() int {
	return 0
}
func (j *JsStub) Index(i int) js.Object {
	return j
}
func (j *JsStub) SetIndex(i int, value interface{}) {}
func (j *JsStub) Call(name string, args ...interface{}) js.Object {
	return j
}
func (j *JsStub) Invoke(args ...interface{}) js.Object {
	return j
}
func (j *JsStub) New(args ...interface{}) js.Object {
	return j
}
func (j *JsStub) Bool() bool {
	return true
}
func (j *JsStub) Str() string {
	return ""
}
func (j *JsStub) Int() int {
	return 0
}
func (j *JsStub) Int64() int64 {
	return 0
}
func (j *JsStub) Uint64() uint64 {
	return 0
}
func (j *JsStub) Float() float64 {
	return 0
}
func (j *JsStub) Interface() interface{} {
	return j
}
func (j *JsStub) Unsafe() uintptr {
	return uintptr(0)
}
func (j *JsStub) IsUndefined() bool {
	return true
}
func (j *JsStub) IsNull() bool {
	return true
}

func stubInit() {
	js.Global = &JsStub{}
}

type Test struct {
	ok bool
}

func TestPromise(t *testing.T) {
	stubInit()
	model := &Test{ok: false}
	p := NewPromise(model, http.Service().NewRequest(http.MethodGet, "/").DoAsync())
	hdler := func(r *http.Response) ModelUpdater {
		return func(m *Test) {
			m.ok = true
		}
	}
	p.OnSuccess(hdler)
	p.fakeResolve(hdler)
	if !model.ok {
		t.Fatalf("The promise handler and/or the updater have not been called.")
	}
	ehdler := func(r *http.Response) ModelUpdater {
		return func(a int) {
		}
	}
	p.OnSuccess(ehdler)
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("There was no panic when calling the model updater of wrong type.")
		}
	}()
	p.fakeResolve(ehdler)
}

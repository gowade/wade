package wade

import (
	"fmt"
	"reflect"

	"github.com/phaikawl/wade/services/http"
)

type Promise struct {
	model interface{}
	d     http.Deferred
}

func NewPromise(model interface{}, op http.Deferred) *Promise {
	return &Promise{model, op}
}

func (p *Promise) Model() interface{} {
	return p.model
}

//Model updater type assertion
func (p *Promise) muTypeAssert(fn ModelUpdater) {
	tpFn := reflect.TypeOf(fn)
	if tpFn.Kind() != reflect.Func {
		panic("The return of promise handler must be a function.")
	}
	paramType := tpFn.In(0).Name()
	if paramType == "" {
		paramType = tpFn.In(0).Elem().Name()
	}
	modelType := reflect.TypeOf(p.model).Name()
	if modelType == "" {
		modelType = reflect.TypeOf(p.model).Elem().Name()
	}
	if paramType != modelType {
		panic(fmt.Sprintf(`The parameter of the promise modelUpdater function (now has type %v) must be of the same type as
		the promise's model (of type %v).`, paramType, modelType))
	}
}

type ModelUpdater interface{}
type PromiseHandlerFunc func(data *http.Response) ModelUpdater

func (p *Promise) handle(fn PromiseHandlerFunc) http.HttpDoneHandler {
	return func(r *http.Response) {
		mu := fn(r)
		p.muTypeAssert(mu)
		reflect.ValueOf(mu).Call([]reflect.Value{reflect.ValueOf(p.model)})
	}
}

func (p *Promise) OnSuccess(fn PromiseHandlerFunc) {
	p.d.Done(p.handle(fn))
}

func (p *Promise) OnFail(fn PromiseHandlerFunc) {
	p.d.Fail(p.handle(fn))
}

func (p *Promise) OnComplete(fn PromiseHandlerFunc) {
	p.d.Then(p.handle(fn))
}

//Manually resolve the promise
func (p *Promise) Resolve() {
	p.d.Resolve()
}

func (p *Promise) fakeResolve(fn PromiseHandlerFunc) {
	(p.handle(fn))(http.NewResponse("", ""))
}

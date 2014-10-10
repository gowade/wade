package bind

import (
	"fmt"
	"reflect"

	"github.com/gopherjs/gopherjs/js"
)

type (
	JsWatchCb func(string, string, js.Object, js.Object)

	WatchCallback func(interface{})
	ReplCallback  func(oldAddr uintptr, repl interface{})
	EvalFn        func(oldAddr uintptr, repl interface{}) interface{}

	ObserveCallback func(oldVal, newVal interface{})

	WatchBackend interface {
		Watch(watchCtl WatchCtl, callback ReplCallback) WatchCloser
		DigestAll(watcher *Watcher)
		Checkpoint()
	}

	WatchCloser interface {
		Close()
	}

	observer struct {
		Callback WatchCallback
		Closer   WatchCloser
		efn      EvalFn
		obj      *ObjEval
		val      interface{}
	}

	Watcher struct {
		backend   WatchBackend
		observers map[uintptr][]*observer
	}

	WatchCtl struct {
		Obj *ObjEval
		obs *observer
		w   *Watcher
	}

	BasicWatchBackend struct{}

	BasicWatchCloser struct {
		watcher *Watcher
		value   reflect.Value
	}
)

func (c WatchCtl) WatchAdd(newFr reflect.Value, closer WatchCloser, callback WatchCallback) {
	_, ok := c.w.observers[newFr.UnsafeAddr()]
	if !ok {
		c.w.observers[newFr.UnsafeAddr()] = []*observer{}
	}

	c.w.observers[newFr.UnsafeAddr()] = append(c.w.observers[newFr.UnsafeAddr()],
		&observer{callback, closer, c.obs.efn, c.obs.obj, nil})
}

func (w WatchCtl) NewFieldRefl() reflect.Value {
	v, ok, err := getReflectField(w.Obj.ModelRefl, w.Obj.Field)
	if !ok || err != nil {
		fmt.Printf("Getting new value for field %v failed.", w.Obj.Field)
	}

	return v
}

func NewWatcher(wb WatchBackend) *Watcher {
	return &Watcher{
		backend:   wb,
		observers: make(map[uintptr][]*observer),
	}
}

func (c BasicWatchCloser) Close() {
	delete(c.watcher.observers, c.value.UnsafeAddr())
}

func (w BasicWatchBackend) Watch(wc WatchCtl, callback ReplCallback) WatchCloser {
	return BasicWatchCloser{wc.w, wc.Obj.FieldRefl}
}

func (w BasicWatchBackend) DigestAll(watcher *Watcher) {
	for _, l := range watcher.observers {
		for _, obs := range l {
			newValue := obs.efn(0, nil)
			rv := reflect.ValueOf(newValue)
			//fmt.Printf("%v %v %v\n", obs, obs.obj.Field, newValue, obs.val)
			if comp, _ := compareRefl(reflect.ValueOf(obs.val), rv); comp != 0 {
				obs.Callback(newValue)
				obs.val = newValue
			}
		}
	}
}

func (w BasicWatchBackend) Checkpoint() {}

func (b *Watcher) Watch(value interface{}, efn EvalFn, obj *ObjEval, callback WatchCallback) {
	obs := &observer{callback, nil, efn, obj, value}
	rcb := func(oldAddr uintptr, repl interface{}) {
		newValue := efn(oldAddr, repl)
		obs.val = newValue

		//gopherjs:blocking
		callback(newValue)
	}

	closer := b.backend.Watch(WatchCtl{obj, obs, b}, rcb)
	obs.Closer = closer

	pt := obj.FieldRefl.UnsafeAddr()
	_, ok := b.observers[pt]
	if !ok {
		b.observers[pt] = make([]*observer, 0)
	}

	b.observers[pt] = append(b.observers[pt], obs)

	return
}

func (b *Watcher) Observe(model interface{}, field string, callback ObserveCallback) (ok bool) {
	oe, ok, err := evaluateObjField(field, reflect.ValueOf(model))
	if err != nil {
		panic(err)
	}

	if !ok {
		return
	}

	old := oe.FieldRefl.Interface()

	b.Watch(old, func(_ uintptr, _ interface{}) interface{} {
		noe, _, _ := evaluateObjField(field, reflect.ValueOf(model))
		return noe.FieldRefl.Interface()
	}, oe, func(newVal interface{}) {
		callback(old, newVal)
		old = newVal
	})

	return
}

// Digest manually triggers the observers for the given object.
// It must be a pointer, normally a pointer to a struct field.
func (b *Watcher) Digest(ptr interface{}) {
	p := reflect.ValueOf(ptr)
	if p.Kind() != reflect.Ptr {
		panic("Argument to ApplyChanges must be a pointer.")
	}

	if p.IsNil() {
		panic("Call of ApplyChanges with nil pointer.")
	}

	for _, ob := range b.observers[p.Elem().UnsafeAddr()] {
		ob.Callback(p.Elem().Interface())
	}
}

func (b *Watcher) Apply(fn func()) {
	fn()
}

func (b *Watcher) Checkpoint() {
	b.backend.Checkpoint()
}

func (b *Watcher) DigestAll() {
	b.backend.DigestAll(b)
}

func (b *Watcher) ResetWatchers() {
	for _, list := range b.observers {
		for _, ob := range list {
			ob.Closer.Close()
		}
	}

	b.observers = make(map[uintptr][]*observer)
}

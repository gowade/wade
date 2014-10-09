package bind

import (
	"fmt"
	"reflect"

	"github.com/gopherjs/gopherjs/js"
)

type (
	JsWatchCb func(string, string, js.Object, js.Object)

	WatchCallback func(uintptr, interface{})

	ObserveCallback func(oldVal, newVal interface{})

	WatchBackend interface {
		Watch(watchCtl WatchCtl, callback WatchCallback) WatchCloser
		DigestAll(watcher *Watcher)
		Checkpoint()
	}

	WatchCloser interface {
		Close()
	}

	observer struct {
		Callback WatchCallback
		Closer   WatchCloser
	}

	Watcher struct {
		backend   WatchBackend
		observers map[uintptr][]observer
	}

	WatchCtl struct {
		ModelRefl reflect.Value
		FieldRefl reflect.Value
		Field     string

		w *Watcher
	}

	BasicWatchBackend struct{}

	BasicWatchCloser struct {
		watcher *Watcher
		value   reflect.Value
	}
)

func (c WatchCtl) WatchAdd(newFr reflect.Value, obs WatchCloser, callback WatchCallback) {
	_, ok := c.w.observers[newFr.UnsafeAddr()]
	if !ok {
		c.w.observers[newFr.UnsafeAddr()] = []observer{}
	}

	c.w.observers[newFr.UnsafeAddr()] = append(c.w.observers[newFr.UnsafeAddr()], observer{callback, obs})
}

func (w WatchCtl) NewFieldRefl() reflect.Value {
	v, ok, err := getReflectField(w.ModelRefl, w.Field)
	if !ok || err != nil {
		fmt.Printf("Getting new value for field %v failed.", w.Field)
	}

	return v
}

func NewWatcher(wb WatchBackend) *Watcher {
	return &Watcher{
		backend:   wb,
		observers: make(map[uintptr][]observer),
	}
}

func (c BasicWatchCloser) Close() {
	delete(c.watcher.observers, c.value.UnsafeAddr())
}

func (w BasicWatchBackend) Watch(wc WatchCtl, callback WatchCallback) WatchCloser {
	return BasicWatchCloser{wc.w, wc.FieldRefl}
}

func (w BasicWatchBackend) DigestAll(watcher *Watcher) {
	for _, l := range watcher.observers {
		for _, obs := range l {
			obs.Callback(0, nil)
		}
	}
}

func (w BasicWatchBackend) Checkpoint() {}

func (b *Watcher) Watch(fieldRefl reflect.Value, modelRefl reflect.Value, field string, callback WatchCallback) {
	closer := b.backend.Watch(WatchCtl{modelRefl, fieldRefl, field, b}, callback)

	pt := fieldRefl.UnsafeAddr()
	_, ok := b.observers[pt]
	if !ok {
		b.observers[pt] = make([]observer, 0)
	}

	b.observers[pt] = append(b.observers[pt], observer{callback, closer})

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

	old := oe.fieldRefl.Interface()

	b.Watch(oe.fieldRefl, oe.modelRefl, oe.field, func(_ uintptr, _ interface{}) {
		noe, _, _ := evaluateObjField(field, reflect.ValueOf(model))
		callback(old, noe.fieldRefl.Interface())
		old = noe.fieldRefl.Interface()
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
		ob.Callback(0, reflect.ValueOf(nil))
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

	b.observers = make(map[uintptr][]observer)
}

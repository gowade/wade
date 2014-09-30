package bind

import (
	"fmt"
	"reflect"

	"github.com/gopherjs/gopherjs/js"
)

type (
	JsWatchCb func(string, string, js.Object, js.Object)

	WatchCallback func(uintptr, reflect.Value)

	ObserveCallback func(oldVal, newVal interface{})

	JsWatcher interface {
		Watch(watchCtl WatchCtl, callback WatchCallback) WatchCloser
		DigestAll()
	}

	WatchCloser interface {
		Close()
	}

	observer struct {
		Callback WatchCallback
		Closer   WatchCloser
	}

	Watcher struct {
		jsWatcher JsWatcher
		observers map[uintptr][]observer
	}

	WatchCtl struct {
		ModelRefl reflect.Value
		FieldRefl reflect.Value
		Field     string

		w *Watcher
	}

	NoopJsWatcher   struct{}
	NoopWatchCloser struct{}
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

func NewWatcher(jsWatcher JsWatcher) *Watcher {
	return &Watcher{
		jsWatcher: jsWatcher,
		observers: make(map[uintptr][]observer),
	}
}

func (NoopWatchCloser) Close() {}

func (w NoopJsWatcher) Watch(wc WatchCtl, callback WatchCallback) WatchCloser {
	return NoopWatchCloser{}
}

func (w NoopJsWatcher) DigestAll() {}

// Watch calls Watch.js to watch the object's changes
func (b *Watcher) Watch(fieldRefl reflect.Value, modelRefl reflect.Value, field string, callback WatchCallback) {
	closer := b.jsWatcher.Watch(WatchCtl{modelRefl, fieldRefl, field, b}, callback)

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

	b.Watch(oe.fieldRefl, oe.modelRefl, oe.field, func(_ uintptr, _ reflect.Value) {
		noe, _, _ := evaluateObjField(field, reflect.ValueOf(model))
		callback(old, noe.fieldRefl.Interface())
		old = noe.fieldRefl.Interface()
	})

	return
}

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
	b.Checkpoint()
}

func (b *Watcher) Checkpoint() {
	b.jsWatcher.DigestAll()
}

func (b *Watcher) ResetWatchers() {
	for _, list := range b.observers {
		for _, ob := range list {
			ob.Closer.Close()
		}
	}

	b.observers = make(map[uintptr][]observer)
}

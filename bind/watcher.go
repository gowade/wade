package bind

import (
	"reflect"
)

type (
	JsWatcher interface {
		Watch(modelRefl reflect.Value, field string, callback func())
	}

	Watcher struct {
		jsWatcher JsWatcher
		watchers  map[reflect.Value][]func()
	}

	NoopJsWatcher struct{}
)

func NewWatcher(jsWatcher JsWatcher) *Watcher {
	return &Watcher{
		jsWatcher: jsWatcher,
		watchers:  make(map[reflect.Value][]func()),
	}
}

func (w NoopJsWatcher) Watch(modelRefl reflect.Value, field string, callback func()) {}

// Watch calls Watch.js to watch the object's changes
func (b Watcher) Watch(fieldRefl reflect.Value, modelRefl reflect.Value, field string, callback func()) {
	b.jsWatcher.Watch(modelRefl, field, callback)
	_, ok := b.watchers[fieldRefl]
	if !ok {
		b.watchers[fieldRefl] = make([]func(), 0)
	}

	b.watchers[fieldRefl] = append(b.watchers[fieldRefl], callback)
}

func (b Watcher) ApplyChanges(ptr interface{}) {
	p := reflect.ValueOf(ptr)
	if p.Kind() != reflect.Ptr {
		panic("Argument to ApplyChanges must be a pointer.")
	}
	if p.IsNil() {
		panic("Call of ApplyChanges with nil pointer.")
	}

	for _, fn := range b.watchers[p.Elem()] {
		fn()
	}
}

func (b Watcher) Apply() {
	for _, olist := range b.watchers {
		for _, fn := range olist {
			fn()
		}
	}
}

func (b Watcher) ResetWatchers() {
	b.watchers = make(map[reflect.Value][]func())
}

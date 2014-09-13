package bind

import (
	"reflect"

	"github.com/gopherjs/gopherjs/js"
)

type (
	JsWatchCb func(string, string, js.Object, js.Object)
	JsWatcher interface {
		Watch(modelRefl reflect.Value, field string, callback func())
	}

	Watcher struct {
		jsWatcher       JsWatcher
		watchersFuncMap map[reflect.Value][]func()
	}

	NoopJsWatcher struct{}
)

func NewWatcher(jsWatcher JsWatcher) *Watcher {
	return &Watcher{
		jsWatcher:       jsWatcher,
		watchersFuncMap: make(map[reflect.Value][]func()),
	}
}

func (w NoopJsWatcher) Watch(modelRefl reflect.Value, field string, callback func()) {}

// Watch calls Watch.js to watch the object's changes
func (b *Watcher) Watch(fieldRefl reflect.Value, modelRefl reflect.Value, field string, callback func()) {
	b.jsWatcher.Watch(modelRefl, field, callback)

	_, ok := b.watchersFuncMap[fieldRefl]
	if !ok {
		b.watchersFuncMap[fieldRefl] = make([]func(), 0)
	}

	b.watchersFuncMap[fieldRefl] = append(b.watchersFuncMap[fieldRefl], callback)
}

func (b *Watcher) ApplyChanges(ptr interface{}) {
	p := reflect.ValueOf(ptr)
	if p.Kind() != reflect.Ptr {
		panic("Argument to ApplyChanges must be a pointer.")
	}
	if p.IsNil() {
		panic("Call of ApplyChanges with nil pointer.")
	}

	for _, fn := range b.watchersFuncMap[p.Elem()] {
		fn()
	}
}

func (b *Watcher) Apply() {
	for _, olist := range b.watchersFuncMap {
		for _, fn := range olist {
			fn()
		}
	}
}

func (b *Watcher) ResetWatchers() {
	b.watchersFuncMap = make(map[reflect.Value][]func())
}

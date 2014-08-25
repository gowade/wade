package jsbackend

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"runtime"

	"code.google.com/p/go.net/html"
	"github.com/gopherjs/gopherjs/js"
	"github.com/phaikawl/wade"
	gqdom "github.com/phaikawl/wade/dom/goquery"
	gohttp "github.com/phaikawl/wade/libs/http/serverside"
)

func init() {
	js.Global = wade.NewStubJsValue(nil)
}

func RenderApp(w io.Writer, conf wade.AppConfig, appFn wade.AppFunc, document io.Reader, server http.Handler, request *http.Request) (err error) {
	defer func() {
		if r := recover(); r != nil {
			trace := make([]byte, 1024)
			count := runtime.Stack(trace, true)
			err = fmt.Errorf("Error while rendering the app: %s\nStack of %d bytes: %s\n", r, count, trace)
		}
	}()

	sourcebytes, err := ioutil.ReadAll(document)
	if err != nil {
		return
	}

	doc := gqdom.GetDom().NewDocument(string(sourcebytes[:]))
	wade.StartApp(conf, appFn, wade.RenderBackend{
		JsBackend: &JsBackend{
			history: wade.NewNoopHistory(),
		},
		Document:    doc,
		HttpBackend: gohttp.ServerBackend{server, request},
	})
	wade.AppServices.PageManager.RedirectToUrl(request.URL.Path)

	err = html.Render(w, doc.(gqdom.Selection).Nodes[0])
	return
}

type (
	JsBackend struct {
		history wade.History
	}

	storage struct {
		values map[string]interface{}
	}
)

func newStorage() storage {
	return storage{make(map[string]interface{})}
}

func (s storage) Get(key string) (v interface{}, ok bool) {
	v, ok = s.values[key]
	return
}

func (s storage) Set(key string, v interface{}) {
	s.values[key] = v
}

func (b *JsBackend) CheckJsDep(symbol string) bool {
	return true
}

// Watch calls Watch.js to watch the object's changes
func (b *JsBackend) Watch(modelRefl reflect.Value, field string, callback func()) {
}

func (b *JsBackend) History() wade.History {
	return b.history
}

func (b *JsBackend) WebStorages() (wade.Storage, wade.Storage) {
	return wade.Storage{newStorage()}, wade.Storage{newStorage()}
}

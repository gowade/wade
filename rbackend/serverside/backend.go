package jsbackend

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"runtime"
	"strings"

	"code.google.com/p/go.net/html"
	"github.com/gopherjs/gopherjs/js"
	"github.com/phaikawl/wade"
	"github.com/phaikawl/wade/bind"
	gqdom "github.com/phaikawl/wade/dom/goquery"
	wadehttp "github.com/phaikawl/wade/libs/http"
	gohttp "github.com/phaikawl/wade/libs/http/serverside"
)

type (
	serverCacheHttpBackend struct {
		gohttp.ServerBackend
		cache       map[string]wadehttp.HttpRecord
		cachePrefix string
	}
)

func (b *serverCacheHttpBackend) Do(r *wadehttp.Request) (err error) {
	err = b.ServerBackend.Do(r)
	if strings.HasPrefix(r.Url, b.cachePrefix) {
		b.cache[wadehttp.RequestIdent(r)] = wadehttp.HttpRecord{r.Response, err}
	}

	return
}

func init() {
	js.Global = wade.NewStubJsValue(nil)
}

func RenderApp(w io.Writer, conf wade.AppConfig, appFn wade.AppFunc, document io.Reader, server http.Handler, request *http.Request, cachePrefix string) (err error) {
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

	cacheb := &serverCacheHttpBackend{
		ServerBackend: gohttp.ServerBackend{server, request},
		cache:         make(map[string]wadehttp.HttpRecord),
		cachePrefix:   cachePrefix,
	}

	doc := gqdom.GetDom().NewDocument(string(sourcebytes[:]))
	wade.StartApp(conf, appFn, wade.RenderBackend{
		JsBackend: &JsBackend{
			NoopJsWatcher: bind.NoopJsWatcher{},
			history:       wade.NewNoopHistory(request.URL.Path),
		},
		Document:    doc,
		HttpBackend: cacheb,
	})

	head := doc.Children().Filter("head")
	if head.Length() == 0 {
		head = doc.NewFragment("<head></head>")
		doc.Prepend(head)
	}
	src := doc.NewFragment(`<script type="text/wadehttp"></script>`)
	cbytes, err := json.Marshal(cacheb.cache)
	if err != nil {
		return
	}

	src.SetHtml(string(cbytes[:]))
	head.Append(src)

	err = html.Render(w, doc.(gqdom.Selection).Nodes[0])
	return
}

type (
	JsBackend struct {
		bind.NoopJsWatcher
		history wade.History
	}

	storage struct {
		values map[string]interface{}
	}
)

func newStorage() storage {
	return storage{make(map[string]interface{})}
}

func (s storage) Get(key string, v interface{}) (ok bool) {
	val, ok := s.values[key]
	if ok {
		reflect.ValueOf(v).Elem().Set(reflect.ValueOf(val))
	}
	return
}

func (s storage) Delete(key string) {
	delete(s.values, key)
}

func (s storage) Set(key string, v interface{}) {
	s.values[key] = v
}

func (b *JsBackend) CheckJsDep(symbol string) bool {
	return true
}

func (b *JsBackend) History() wade.History {
	return b.history
}

func (b *JsBackend) WebStorages() (wade.Storage, wade.Storage) {
	return wade.Storage{newStorage()}, wade.Storage{newStorage()}
}

package serverside

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"runtime"
	"strings"

	gqdom "github.com/phaikawl/wade/dom/goquery"
	"golang.org/x/net/html"

	"github.com/phaikawl/wade"
	"github.com/phaikawl/wade/app"
	"github.com/phaikawl/wade/dom"
	wadehttp "github.com/phaikawl/wade/libs/http"
	gohttp "github.com/phaikawl/wade/libs/http/serverside"
)

type (
	serverCacheHttpBackend struct {
		gohttp.ServerBackend
		cache       map[string]*requestList
		cachePrefix string
	}

	requestList struct {
		Records []wadehttp.HttpRecord
	}

	Backend struct {
		history     wade.History
		httpBackend *serverCacheHttpBackend
		document    dom.Selection
	}
)

func (b Backend) History() wade.History {
	return b.history
}

func (b Backend) Bootstrap(app *app.Application) {}

func (b Backend) Document() dom.Selection {
	return b.document
}

func (b Backend) HttpBackend() wadehttp.Backend {
	return b.httpBackend
}

func (b *serverCacheHttpBackend) Do(r *wadehttp.Request) (err error) {
	err = b.ServerBackend.Do(r)
	if strings.HasPrefix(r.URL.String(), b.cachePrefix) {
		rid := wadehttp.RequestIdent(r)
		if _, ok := b.cache[rid]; !ok {
			b.cache[rid] = &requestList{make([]wadehttp.HttpRecord, 0)}
		}

		b.cache[rid].Records = append(b.cache[rid].Records, wadehttp.HttpRecord{r.Response, err})
	}

	return
}

func RenderApp(
	server http.Handler,
	request *http.Request,
	w io.Writer,
	conf app.Config,
	appMain app.Main,
	document io.Reader,
	cachePrefix string) (err error) {

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
		cache:         make(map[string]*requestList),
		cachePrefix:   cachePrefix,
	}

	doc := gqdom.GetDom().NewDocument(string(sourcebytes[:]))
	app := app.New(conf, Backend{
		history:     wade.NewNoopHistory(request.URL.Path),
		document:    doc,
		httpBackend: cacheb,
	})

	app.Start(appMain)

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

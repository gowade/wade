package serverside

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
	"strings"

	gqdom "github.com/phaikawl/wade/dom/gonet"
	"golang.org/x/net/html"

	"github.com/phaikawl/wade/app"
	"github.com/phaikawl/wade/dom"
	wadehttp "github.com/phaikawl/wade/libs/http"
	gohttp "github.com/phaikawl/wade/libs/http/serverside"
	"github.com/phaikawl/wade/page"
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

	renderBackend struct {
		history     page.History
		httpBackend wadehttp.Backend
		document    dom.Selection
	}
)

func (b renderBackend) History() page.History {
	return b.history
}

func (b renderBackend) Bootstrap(app *app.Application) {}

func (b renderBackend) Document() dom.Selection {
	return b.document
}

func (b renderBackend) HttpBackend() wadehttp.Backend {
	return b.httpBackend
}

func (b renderBackend) AfterReady(app *app.Application) {
	if bkn, ok := b.httpBackend.(*serverCacheHttpBackend); ok {
		bkn.AfterReady(app)
	}
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

func (b *serverCacheHttpBackend) requestPath() string {
	return b.ClientReq.URL.Path
}

func (b *serverCacheHttpBackend) AfterReady(app *app.Application) {
	doc := app.Document()
	head := doc.Children().Filter("head")
	if head.Length() == 0 {
		head = doc.NewFragment("<head></head>")
		doc.Prepend(head)
	}

	src := doc.NewFragment(`<script type="text/wadehttp"></script>`)
	cbytes, err := json.Marshal(b.cache)
	if err != nil {
		return
	}

	src.SetHtml(string(cbytes[:]))
	head.Append(src)
}

func NewHttpBackend(server http.Handler, request *http.Request, cachePrefix string) wadehttp.Backend {
	return &serverCacheHttpBackend{
		ServerBackend: gohttp.ServerBackend{server, request},
		cache:         make(map[string]*requestList),
		cachePrefix:   cachePrefix,
	}
}

func NewApp(conf app.Config, document io.Reader, startPath string, httpBackend wadehttp.Backend) *app.Application {
	sourcebytes, err := ioutil.ReadAll(document)
	if err != nil {
		log.Println(`HTML parse error "%v".`, err.Error())
	}

	doc := gqdom.GetDom().NewDocument(string(sourcebytes[:]))

	return app.New(conf, renderBackend{
		history:     page.NewNoopHistory(startPath),
		document:    doc,
		httpBackend: httpBackend,
	})
}

func StartRender(app *app.Application, appMain app.Main, w io.Writer) (err error) {
	defer func() {
		if r := recover(); r != nil {
			trace := make([]byte, 4096)
			count := runtime.Stack(trace, true)
			err = fmt.Errorf("Error while starting and rendering the app: %s\nStack of %d bytes: %s\n", r, count, trace)
		}
	}()

	err = app.Start(appMain)
	if err != nil {
		return
	}

	err = html.Render(w, app.Document().(gqdom.Selection).Nodes[0])
	return
}

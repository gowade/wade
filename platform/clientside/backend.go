package clientside

import (
	"encoding/json"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"github.com/phaikawl/wade/app"
	"github.com/phaikawl/wade/dom"
	"github.com/phaikawl/wade/dom/jsdom"
	"github.com/phaikawl/wade/libs/http"
	xhr "github.com/phaikawl/wade/libs/http/clientside"
	"github.com/phaikawl/wade/page"
)

var (
	gJQ               = jquery.NewJQuery
	gGlobal js.Object = js.Global
)

type (
	cachedHttpBackend struct {
		http.Backend
		cache map[string]*requestList
	}

	headers struct {
		Header http.HttpHeader
	}

	concreteResponse struct {
		http.Response
		Headers headers
	}

	concreteRecord struct {
		Response *concreteResponse
		http.HttpRecord
	}

	requestList struct {
		Records []concreteRecord
		index   int
	}

	renderBackend struct {
		history     page.History
		httpBackend http.Backend
		document    dom.Selection
	}
)

func (b renderBackend) History() page.History {
	return b.history
}

func (b renderBackend) Bootstrap(app *app.Application) {
	err := jsDepCheck()
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			<-app.EventFinished()
			app.Render()
		}
	}()
}

func (b renderBackend) Document() dom.Selection {
	return b.document
}

func (b renderBackend) HttpBackend() http.Backend {
	return b.httpBackend
}

func (b renderBackend) AfterReady(ap *app.Application) {
}

func CreateBackend() renderBackend {
	doc := jsdom.Document()

	return renderBackend{
		history:     History{js.Global.Get("history")},
		document:    doc,
		httpBackend: newCachedHttpBackend(xhr.XhrBackend{}, doc),
	}
}

func (r *requestList) Pop() (re concreteRecord) {
	re = r.Records[r.index]
	r.index++
	return
}

func newCachedHttpBackend(backend http.Backend, doc dom.Selection) *cachedHttpBackend {
	b := &cachedHttpBackend{backend, make(map[string]*requestList)}
	sn := doc.Find("script[type='text/wadehttp']")
	if sn.Length() > 0 {
		cc := sn.Text()
		if cc != "" {
			err := json.Unmarshal([]byte(cc), &b.cache)
			if err != nil {
				panic(err.Error())
			}
		}
	}

	return b
}

func (c *cachedHttpBackend) Do(r *http.Request) (err error) {
	if list, ok := c.cache[http.RequestIdent(r)]; ok && list.index < len(list.Records) {
		record := list.Pop()
		err = record.Error
		r.Response = &record.Response.Response
	} else {
		//gopherjs:blocking
		err = c.Backend.Do(r)
	}

	return
}

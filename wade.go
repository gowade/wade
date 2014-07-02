package wade

import (
	"fmt"

	"github.com/gopherjs/gopherjs/js"
	jq "github.com/gopherjs/jquery"
	"github.com/phaikawl/wade/services/http"
)

var (
	gHistory    js.Object
	gJQ         = jq.NewJQuery
	WadeDevMode = true
)

const (
	AttrPrefix = "attr-"
	BindPrefix = "bind-"
)

type Wade struct {
	pm         *PageManager
	tm         *CustagMan
	tcontainer jq.JQuery
	binding    *Binding
}

func WadeUp(startPage, basePath string, tempcontainer, container string, initFn func(*Wade)) *Wade {
	gHistory = js.Global.Get("history")
	origin := js.Global.Get("document").Get("location").Get("origin").Str()
	tempContainer := gJQ("script[type='text/wadin']#" + tempcontainer)
	if tempContainer.Length == 0 {
		panic(fmt.Sprintf("Template container #%v not found or is wrong kind of element, must be script[type='text/wadin'].",
			tempContainer))
	}
	html := js.Global.Get(jq.JQ).Call("parseHTML", "<div>"+tempContainer.Html()+"</div>")
	tElem := gJQ(html)
	htmlImport(tElem, origin)
	tm := newCustagMan(tElem)
	binding := newBindEngine(tm)
	wd := &Wade{
		pm:         newPageManager(startPage, basePath, container, tElem, binding, tm),
		tm:         tm,
		binding:    binding,
		tcontainer: tElem,
	}
	wd.Init()
	initFn(wd)
	return wd
}

func (wd *Wade) Pager() *PageManager {
	return wd.pm
}

func (wd *Wade) Custags() *CustagMan {
	return wd.tm
}

func (wd *Wade) Binding() *Binding {
	return wd.binding
}

func htmlImport(parent jq.JQuery, origin string) {
	parent.Find("wimport").Each(func(i int, elem jq.JQuery) {
		src := elem.Attr("src")
		req := http.NewRequest(http.MethodGet, origin+src)
		html := req.DoSync().Data()
		elem.Append(html)
		htmlImport(elem, origin)
	})
}

type UrlInfo struct {
	path    string
	fullUrl string
}

func (wd *Wade) Init() {
	wd.binding.RegisterHelper("url", func(pageid string, params ...interface{}) UrlInfo {
		url, err := wd.pm.PageUrl(pageid, params)
		if err != nil {
			panic(fmt.Errorf(`url helper error: "%v", when getting url for page "%v"`, err.Error(), pageid))
		}
		return UrlInfo{url, wd.pm.Url(url)}
	})
}

func (wd *Wade) Start() {
	gJQ(js.Global.Get("document")).Ready(func() {
		wd.tm.prepare()
		wd.pm.getReady()
	})
}

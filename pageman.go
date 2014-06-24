package wade

import (
	"fmt"
	"strings"

	"github.com/gopherjs/gopherjs/js"
	jq "github.com/gopherjs/jquery"
)

type pageInfo struct {
	path  string
	title string
}

type PageController func() interface{}
type PageHandler func()

type PageManager struct {
	router          js.Object
	currentPage     string
	pageHandlers    map[string][]PageHandler
	pageControllers map[string]PageController
	startPage       string
	basePath        string
	pages           map[string]pageInfo
	notFoundPage    string
	container       jq.JQuery
	tempElem        jq.JQuery
	//pageModels   []js.Object
}

func newPageManager(startPage, basePath string, container string, tempElem jq.JQuery) *PageManager {
	return &PageManager{
		router:          js.Global.Get("RouteRecognizer").New(),
		currentPage:     "",
		pageHandlers:    make(map[string][]PageHandler),
		pageControllers: make(map[string]PageController),
		basePath:        basePath,
		startPage:       startPage,
		pages:           make(map[string]pageInfo),
		notFoundPage:    "",
		container:       gJQ("#" + container),
		tempElem:        tempElem,
		//pageModels:   make([]js.Object, 0),
	}
}

func (pm *PageManager) cutPath(path string) string {
	if strings.HasPrefix(path, pm.basePath) {
		path = path[len(pm.basePath):]
	}
	return path
}

func (pm *PageManager) page(pageId string) pageInfo {
	if page, ok := pm.pages[pageId]; ok {
		return page
	}
	panic(fmt.Sprintf("no such page #%v found.", pageId))
}

func (pm *PageManager) SetNotFoundPage(pageId string) {
	_ = pm.page(pageId)
	pm.notFoundPage = pageId
}

func (pm *PageManager) Url(path string) string {
	return pm.basePath + path
}

func documentUrl() string {
	location := gHistory.Get("location")
	if location.IsNull() || location.IsUndefined() {
		location = js.Global.Get("document").Get("location")
	}
	return location.Get("pathname").Str()
}

func elemForPage(parent jq.JQuery, pageid string) jq.JQuery {
	elem := parent.Find("#" + pageid)
	if elem.Length == 0 {
		panic(fmt.Sprintf("Cannot find element for page #%v.", pageid))
	}
	if WadeDevMode {
		elem = parent.Find("[id='" + pageid + "']")
		if elem.Length > 1 {
			panic(fmt.Sprintf("Unexpected: duplicated element id #%v when getting the page element.", pageid))
		}
		if !elem.Is("wpage") {
			panic(fmt.Sprintf("page element #%v must be a wpage.", pageid))
		}
	}
	return elem
}

func (pm *PageManager) setupPageOnLoad() {
	path := pm.cutPath(documentUrl())
	if path == "/" {
		path = pm.page(pm.startPage).path
		gHistory.Call("replaceState", nil, pm.pages[pm.startPage].title, pm.Url(path))
	}
	pm.updatePage(path)
}

func (pm *PageManager) getReady() {
	if pm.container.Length == 0 {
		panic(fmt.Sprintf("Cannot find the page container #%v.", pm.container))
	}

	pm.tempElem.Find("a").Each(func(idx int, a jq.JQuery) {
		href := a.Attr("href")
		if strings.HasPrefix(href, ":") {
			pageId := string([]rune(href)[1:])
			a.SetAttr("href", pm.Url(pm.page(pageId).path))
			a.SetAttr("data-wade-page", pageId)
		}
	})

	gJQ(js.Global.Get("window")).On("popstate", func() {
		pm.updatePage(documentUrl())
	})

	pm.setupPageOnLoad()
}

func (pm *PageManager) updatePage(url string) {
	url = pm.cutPath(url)
	matches := pm.router.Call("recognize", url)
	println("path: " + url)
	if matches.IsUndefined() || matches.Length() == 0 {
		if pm.notFoundPage != "" {
			pm.updatePage(pm.page(pm.notFoundPage).path)
		} else {
			panic("Page not found. No 404 handler declared.")
		}
	}

	pageId := matches.Index(0).Get("handler").Invoke().Str()

	pageElem := elemForPage(pm.tempElem, pageId)
	gJQ("title").SetText(pm.page(pageId).title)
	if pm.currentPage != pageId {
		jqparents := pageElem.Parents("wpage")
		leng := jqparents.Length
		parents := make([]jq.JQuery, leng+1)
		resultElems := make([]jq.JQuery, leng)
		for i := 0; i < leng; i++ {
			parents[i] = jqparents.Eq(leng - i - 1)
			clone := gJQ(parents[i].Get(0).Call("cloneNode"))
			resultElems[i] = clone
			if i > 0 {
				resultElems[i-1].Append(clone)
			}
		}
		parents[leng] = pageElem
		for i := leng - 1; i >= 0; i-- {
			p := parents[i]
			p.Children("*").Each(func(_ int, e jq.JQuery) {
				if !e.Is("wpage") || e.Is(parents[i+1].Get(0)) {
					resultElems[i].Append(e.Clone())
				}
			})
		}

		pm.container.SetHtml(resultElems[0].Html())

		//Rebind link events
		pm.container.Find("a").On(jq.CLICK, func(e jq.Event) {
			a := gJQ(e.Target)

			pageId := a.Attr("data-wade-page")
			if pageId == "" { //not a wade page link, let the browser do its job
				return
			}

			e.PreventDefault()

			pageInf := pm.page(pageId)
			gHistory.Call("pushState", nil, pageInf.title, pm.Url(pageInf.path))
			pm.updatePage(pageInf.path)
		})

		pm.currentPage = pageId
	}
}

func (pm *PageManager) bindPage(b *binding) {
	if handlers, ok := pm.pageHandlers[pm.currentPage]; ok {
		for _, handler := range handlers {
			handler()
		}
	}
	if controller, exist := pm.pageControllers[pm.currentPage]; exist {
		model := controller()
		b.Bind(elemForPage(pm.container, pm.currentPage), model)
	}
}

func (pm *PageManager) RegisterController(pageId string, fn PageController) {
	if _, exist := pm.pageControllers[pageId]; exist {
		panic(fmt.Sprintf("That page #%v already has a controller.", pageId))
	}
	pm.pageControllers[pageId] = fn
}

func (pm *PageManager) RegisterHandler(pageId string, fn PageHandler) {
	if _, exist := pm.pageHandlers[pageId]; !exist {
		pm.pageHandlers[pageId] = make([]PageHandler, 0)
	}
	pm.pageHandlers[pageId] = append(pm.pageHandlers[pageId], fn)
}

func (pm *PageManager) RegisterPages(pages map[string]string) {
	for path, pageId := range pages {
		if _, exist := pm.pages[pageId]; exist {
			panic(fmt.Sprintf("Page #%v has already been registered.", pageId))
		}
		pageElem := pm.tempElem.Find("#" + pageId)
		if pageElem.Length == 0 {
			panic(fmt.Sprintf("There is no such page element #%v.", pageId))
		}

		(func(path, pageId string) {
			pm.router.Call("add", []map[string]interface{}{
				map[string]interface{}{
					"path": path,
					"handler": func() string {
						return pageId
					},
				},
			})
		})(path, pageId)

		pm.pages[pageId] = pageInfo{path: path, title: pageElem.Attr("title")}
	}
}

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

type PageHandler func() interface{}

type PageManager struct {
	router       js.Object
	currentPage  string
	pageHandlers map[string][]PageHandler
	startPage    string
	basePath     string
	pages        map[string]pageInfo
	notFoundPage string
	//pageModels   []js.Object
}

func newPageManager(startPage, basePath string) *PageManager {
	return &PageManager{
		router:       js.Global.Get("RouteRecognizer").New(),
		currentPage:  startPage,
		pageHandlers: make(map[string][]PageHandler),
		basePath:     basePath,
		startPage:    startPage,
		pages:        make(map[string]pageInfo),
		notFoundPage: "",
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

func (pm *PageManager) setupPageOnLoad() {
	path := pm.cutPath(documentUrl())
	if path == "/" {
		path = pm.page(pm.startPage).path
		gHistory.Call("replaceState", nil, pm.pages[pm.startPage].title, pm.Url(path))
	}
	pm.updatePage(path)
}

func (pm *PageManager) getReady() {
	gJQ("a").On(jq.CLICK, func(e jq.Event) {
		e.PreventDefault()
		a := gJQ(e.Target)
		href := a.Attr("href")
		if href == "" || !strings.HasPrefix(href, ":") {
			return
		}

		pageId := string([]rune(href)[1:])
		pageInf := pm.page(pageId)
		gHistory.Call("pushState", nil, pageInf.title, pm.Url(pageInf.path))
		pm.updatePage(pageInf.path)
	})

	gJQ(js.Global.Get("window")).On("popstate", func() {
		pm.updatePage(documentUrl())
	})

	gJQ("welement").Hide()
	pm.setupPageOnLoad()
}

func pageHide(elems jq.JQuery) {
	elems.Hide()
	elems.SetData("hidden", "t")
}

func pageShow(elems jq.JQuery) {
	elems.Show()
	elems.SetData("hidden", "")
}

func pageIsHidden(pageElem jq.JQuery) bool {
	return pageElem.Data("hidden") == "t"
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
	pageElem := gJQ("#" + pageId)
	gJQ("title").SetText(pm.page(pageId).title)
	if pm.currentPage != pageId {
		cpElem := gJQ("#" + pm.currentPage)
		pageHide(cpElem)
		cpElem.Parents("wpage").Each(func(idx int, p jq.JQuery) {
			pageHide(p)
		})
	}

	pageElem.Parents("wpage").Each(func(idx int, p jq.JQuery) {
		pageShow(p)
	})
	pageShow(pageElem)
	pm.currentPage = pageId
}

func (pm *PageManager) bindPage(b *binding) {
	if handlers, ok := pm.pageHandlers[pm.currentPage]; ok {
		for _, handler := range handlers {
			model := handler()
			b.Bind(gJQ("#"+pm.currentPage), model)
			//println(gJQ("#" + pm.currentPage).Underlying().Interface())
		}
	}
}

func (pm *PageManager) inCurrentPage(elem jq.JQuery) bool {
	//return elem.Closest(pm.currentPage).Length > 0
	return !pageIsHidden(elem.Parent("wpage").First())
}

func (pm *PageManager) RegisterHandler(pageId string, handlerFn PageHandler) {
	if _, exist := pm.pageHandlers[pageId]; !exist {
		pm.pageHandlers[pageId] = make([]PageHandler, 0)
	}
	pm.pageHandlers[pageId] = append(pm.pageHandlers[pageId], handlerFn)
}

func (pm *PageManager) RegisterPages(pages map[string]string) {
	for path, pageId := range pages {
		if _, exist := pm.pages[pageId]; exist {
			panic(fmt.Sprintf("Page #%v has already been registered.", pageId))
		}
		pageElem := gJQ("#" + pageId)
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
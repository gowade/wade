package wade

import (
	"fmt"
	"path"
	"strings"

	"github.com/gopherjs/gopherjs/js"
	urlrouter "github.com/naoina/kocha-urlrouter"
	_ "github.com/naoina/kocha-urlrouter/regexp"
	"github.com/phaikawl/wade/bind"
	"github.com/phaikawl/wade/dom"
	"github.com/phaikawl/wade/icommon"
)

const (
	GlobalDisplayScope = "__global__"
)

type (
	bindEngine interface {
		Watcher() *bind.Watcher
		BindModels(root dom.Selection, models []interface{}, once bool)
	}

	History interface {
		ReplaceState(title string, path string)
		PushState(title string, path string)
		OnPopState(fn func())
		CurrentPath() string
		RedirectTo(url string)
	}

	pageManager struct {
		document      dom.Selection
		routes        []urlrouter.Record
		router        urlrouter.URLRouter
		currentPage   *page
		basePath      string
		notFoundPage  *page
		rcProto       string
		container     dom.Selection
		tcontainer    dom.Selection
		realContainer dom.Selection

		binding        bindEngine
		pc             *Scope
		displayScopes  map[string]displayScope
		globalDs       *globalDisplayScope
		formattedTitle string
		history        History
	}
)

func newPageManager(history History, config AppConfig, document dom.Selection,
	tcontainer dom.Selection, binding bindEngine) *pageManager {

	realContainer := config.Container
	if realContainer == nil {
		body := document.Find("body")
		realContainer = body.Find(".wade-app-container")
		if realContainer.Length() > 0 {
			realContainer = realContainer.First()
		} else {
			realContainer = document.NewFragment(`<div class="wade-app-container"></div>`)
			body.Prepend(realContainer)
		}
	}

	if realContainer.Length() == 0 {
		panic("App container doesn't exist.")
	}

	basePath := config.BasePath
	if basePath == "" {
		basePath = "/"
	}

	cl := realContainer.Clone()
	cl.SetHtml("")

	pm := &pageManager{
		document:      document,
		routes:        make([]urlrouter.Record, 0),
		router:        urlrouter.NewURLRouter("regexp"),
		currentPage:   nil,
		basePath:      basePath,
		notFoundPage:  nil,
		rcProto:       cl.OuterHtml(),
		tcontainer:    tcontainer,
		realContainer: realContainer,
		binding:       binding,
		displayScopes: make(map[string]displayScope),
		globalDs:      &globalDisplayScope{},
		history:       history,
	}

	pm.displayScopes[GlobalDisplayScope] = pm.globalDs
	return pm
}

func (pm *pageManager) addRoute(p *page) {
	pm.routes = append(pm.routes, urlrouter.NewRecord(p.path, p))
}

func (pm *pageManager) CurrentPageId() string {
	return pm.currentPage.id
}

func (pm *pageManager) cutPath(path string) string {
	if strings.HasPrefix(path, pm.basePath) {
		path = path[len(pm.basePath):]
	}
	return path
}

func (pm *pageManager) page(id string) *page {
	if ds, hasDs := pm.displayScopes[id]; hasDs {
		if page, ok := ds.(*page); ok {
			return page
		}
	}

	panic(fmt.Sprintf(`No such page "%v"`, id))
}

func (pm *pageManager) SetNotFoundPage(pageId string) {
	pm.notFoundPage = pm.page(pageId)
}

// Url returns the full path
func (pm *pageManager) FullPath(pa string) string {
	return path.Join(pm.basePath, pa)
}

func (pm *pageManager) RedirectToPage(page string, params ...interface{}) {
	url, err := pm.PageUrl(page, params...)
	if err != nil {
		panic(err.Error())
	}

	pm.updatePage(url, true)
}

func (pm *pageManager) RedirectToUrl(url string) {
	if strings.HasPrefix(url, pm.BasePath()) {
		pm.updatePage(url, true)
	} else {
		pm.history.RedirectTo(url)
	}
}

func (pm *pageManager) prepare() {
	//build the router
	pm.router.Build(pm.routes)

	pm.history.OnPopState(func() {
		go func() {
			pm.updatePage(pm.history.CurrentPath(), false)
		}()
	})

	pm.updatePage(pm.history.CurrentPath(), false)
}

func walk(elem dom.Selection, pm *pageManager) {
	for _, e := range elem.Children().Elements() {
		belong, ok := e.Attr("w-belong")
		if !ok {
			walk(e, pm)
		} else {
			if ds, ok := pm.displayScopes[belong]; ok {
				if ds.hasPage(pm.currentPage.id) {
					walk(e, pm)
				} else {
					e.Remove()
				}
			} else {
				panic(fmt.Sprintf(`Invalid value "%v" for w-belong, no such page or page group is registered.`, belong))
			}
		}
	}
}

func newHiddenContainer(rcProto string, document dom.Dom) dom.Selection {
	hiddenRoot := document.NewRootFragment()
	container := document.NewFragment(rcProto)
	hiddenRoot.Append(container)
	return container
}

type scrollItem struct {
	elem dom.Selection
	posx int
	posy int
}

type scrollPreserver struct {
	scrolls []scrollItem
}

func (sp *scrollPreserver) getScroll(oldElem dom.Selection) {
	if posy, posx := oldElem.Underlying().Call("scrollTop").Int(),
		oldElem.Underlying().Call("scrollLeft").Int(); posx != 0 || posy != 0 {
		sp.scrolls = append(sp.scrolls, scrollItem{oldElem, posx, posy})
	}
}

func (sp *scrollPreserver) applyScrolls(newCtn dom.Selection) {
	for _, item := range sp.scrolls {
		ne := dom.GetElemCounterpart(item.elem, newCtn)
		if item.posy != 0 {
			ne.Underlying().Call("scrollTop", item.posy)
		}
		if item.posx != 0 {
			ne.Underlying().Call("scrollLeft", item.posx)
		}
	}
}

func (pm *pageManager) updatePage(url string, pushState bool) {
	path := pm.cutPath(url)

	match, routeparams := pm.router.Lookup(path)
	if match == nil {
		if pm.notFoundPage != nil {
			pm.updatePage(pm.notFoundPage.path, false)
		} else {
			panic("Page not found. No 404 page declared.")
		}
		return
	}

	page := match.(*page)

	if pushState {
		pm.history.PushState(page.title, pm.FullPath(path))
	}

	params := make(map[string]interface{})
	for _, param := range routeparams {
		params[param.Name] = param.Value
	}

	pm.formattedTitle = page.title

	pm.currentPage = page
	pm.container = newHiddenContainer(pm.rcProto, pm.document)
	pm.container.SetHtml(pm.tcontainer.Html())
	pm.container.Find("wdefine").Remove()

	walk(pm.container, pm)

	pm.binding.Watcher().ResetWatchers()
	pm.bind(params)
	icommon.WrapperUnwrap(pm.container)
	pm.setTitle(pm.formattedTitle)

	if js.Global != nil && !js.Global.Get("window").IsUndefined() {
		jqwindow := js.Global.Call("jQuery", js.Global.Get("window"))
		scrollpos := jqwindow.Call("scrollTop")
		sp := &scrollPreserver{[]scrollItem{}}
		for _, c := range pm.realContainer.Find(".w-scrolled").Elements() {
			sp.getScroll(c)
		}
		pm.realContainer.ReplaceWith(pm.container)
		sp.applyScrolls(pm.container)

		jqwindow.Call("scrollTop", scrollpos)
	} else {
		pm.realContainer.ReplaceWith(pm.container)
	}
	pm.realContainer = pm.container

	//Handle link events
	pm.realContainer.Listen("click", "a", func(e dom.Event) {
		href, ok := e.Target().Attr("href")
		if !ok {
			return
		}

		if !strings.HasPrefix(href, pm.BasePath()) { //not a wade page link, let the browser do its job
			return
		}

		e.PreventDefault()

		go func() {
			pm.updatePage(href, true)
		}()
	})
}

func (pm *pageManager) setTitle(title string) {
	tElem := pm.document.NewFragment("<title>" + title + "</title>")
	head := pm.document.Find("head").First()
	oElem := head.Find("title")
	if oElem.Length() == 0 {
		head.Append(tElem)
	} else {
		oElem.ReplaceWith(tElem)
	}
}

// PageUrl returns the url for the page with the given parameters
func (pm *pageManager) PageUrl(pageId string, params ...interface{}) (u string, err error) {
	u, err = pm.pageUrl(pageId, params)
	u = pm.FullPath(u)
	return
}

func (pm *pageManager) pageUrl(pageId string, params []interface{}) (u string, err error) {
	err = nil
	page := pm.page(pageId)

	k, i := 0, 0
	route := page.path
	routeparams := urlrouter.ParamNames(route)
	for {
		if i >= len(route) {
			break
		}

		if urlrouter.IsMetaChar(route[i]) && route[i:i+len(routeparams[k])] == routeparams[k] {
			if k < len(params) && params[k] != nil {
				u += fmt.Sprintf("%v", params[k])
			}
			i += len(routeparams[k])
			k++
		} else {
			u += string(route[i])
			i++
		}
	}

	if k != len(params) || k != len(routeparams) {
		err = fmt.Errorf(`Wrong number of parameters for the route of %v. Expected %v, got %v.`,
			pageId, len(routeparams), len(params))
		return
	}

	return
}

func (pm *pageManager) BasePath() string {
	return pm.basePath
}

func (pm *pageManager) CurrentPage() *Scope {
	return pm.pc
}

func (pm *pageManager) bind(params map[string]interface{}) {
	s := pm.newRootScope(pm.currentPage, params)
	controllers := make([]PageControllerFunc, 0)

	add := func(ds displayScope) {
		if ctrls := ds.Controllers(); ctrls != nil {
			for _, controller := range ctrls {
				controllers = append(controllers, controller)
			}
		}
	}

	add(pm.globalDs)
	for _, grp := range pm.currentPage.groups {
		add(grp)
	}
	add(pm.currentPage)

	if len(controllers) > 0 {
		completeChan := make(chan bool, 1)
		queueChan := make(chan bool, len(controllers))
		for _, controller := range controllers {
			go func(controller PageControllerFunc) {
				//gopherjs:blocking
				err := controller(s)
				if err != nil {
					panic(err)
				}

				queueChan <- true
				if len(queueChan) == len(controllers) {
					completeChan <- true
				}
			}(controller)
		}
		<-completeChan
	}

	//gopherjs:blocking
	pm.binding.BindModels(pm.container, s.bindModels(), false)

	pm.pc = s

	pm.binding.Watcher().Checkpoint()
}

// RegisterController adds a new controller function for the specified
// page / page group.
func (pm *pageManager) registerController(displayScope string, fn PageControllerFunc) {
	ds, ok := pm.displayScopes[displayScope]
	if !ok {
		panic(fmt.Errorf(`Registering controller for "%v", there's no such page or page group.`, displayScope))
	}

	ds.addController(fn)
}

// RegisterDisplayScopes registers the given maps of pages and pageGroups
func (pm *pageManager) registerDisplayScopes(pages []PageDesc, pageGroups []PageGroupDesc) {
	if pages != nil {
		for _, pg := range pages {
			pm.displayScopes[pg.id] = pg.Register(pm)
		}
	}

	if pageGroups != nil {
		for _, pg := range pageGroups {
			pm.displayScopes[pg.id] = pg.Register(pm)
		}
	}
}
